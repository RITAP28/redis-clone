package cache

import (
	"fmt"
	"slices"
	"sync"
)

func(r *RedisCache) IsAlreadySubscribed(channel string, client *Client) bool {
	v := r.pubsubs.channels[channel]
	return slices.Contains(v, client)
}

func(r *RedisCache) removeClientFromAllChannels(c *Client) {
	r.pubsubs.mu.Lock()
	defer r.pubsubs.mu.Unlock()

	if c == nil {
		return
	}

	channels := r.pubsubs.channels
	for chnl := range channels {
		newList := []*Client{}
		clients := channels[chnl]
		for _, client := range clients {
			if c != client {
				newList = append(newList, client)
			}
		}

		if len(newList) == 0 {
			delete(r.pubsubs.channels, chnl)
		} else {
			channels[chnl] = newList
		}
	}

	c.Subscriptions = []string{}
	c.inSubscription = false

	fmt.Printf("Cleaned up disconnected client from all channels\n")
}

func(r *RedisCache) Subscribe(client *Client, channels []string) (int, bool) {
	// the whole redis cache shall not be blocked for executing pubsub commands
	r.pubsubs.mu.Lock()
	defer r.pubsubs.mu.Unlock()

	if len(channels) < 1 {
		return 0, false
	}

	// appending the client to the channel in the global registry
	for _, channel := range channels {
		// manual checking whether channel exists or not through loops is not needed
		// go maps already handle that case
		if !r.IsAlreadySubscribed(channel, client) {
			client.Subscriptions = append(client.Subscriptions, channel)
			r.pubsubs.channels[channel] = append(r.pubsubs.channels[channel], client)
		}
	}

	// turning the subscription mode ON for this client
	client.inSubscription = true
	activeSubscriptions := len(client.Subscriptions)

	// returns true and number of active subscriptions the client has now
	return activeSubscriptions, true
}

func(r *RedisCache) Publish(channel string, message string) (int, bool) {
	// prevents blocking normal redis operations (SET, GET, DELETE)
	// multiple publishers can publish concurrently
	r.pubsubs.mu.RLock()
	subscribers, exists := r.pubsubs.channels[channel]
	if !exists || len(subscribers) == 0 {
		return 0, false
	}

	// making a copy of the clients to prevent panic if someone unsubscribes while the messages are being sent
	clients := append([]*Client(nil), subscribers...)
	r.pubsubs.mu.RUnlock()

	count := 0
	var wg sync.WaitGroup

	// getting all the clients subscribed to the mentioned channel
	formattedMessage := fmt.Sprintf("*3\r\n$7\r\nmessage\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(channel), channel, len(message), message)

	// message shall be sent to all the clients, including the publisher
	// spawning goroutines for each client to prevent blocking the publisher if any subscriber is slow and handle a large number of client connections at once
	var mu sync.Mutex
	for _, client := range clients {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()

			_, err := client.Conn.Write([]byte(formattedMessage))
			if err != nil {
				// the below warning means there are some clients inside the global registry of channels which are not functioning or dead
				// so, it's good if they are removed from all channels (cleanup)
				fmt.Printf("Warning: Failed to send message to client: %v\n", err)
				r.removeClientFromAllChannels(client)
				return
			}

			mu.Lock()
			count++
			mu.Unlock()
		}(client)
	}

	wg.Wait()
	return count, true
}

func(r *RedisCache) Unsubscribe(client *Client, channels []string) (string, bool) {
	r.pubsubs.mu.Lock()
	defer r.pubsubs.mu.Unlock()

	// if no channels mentioned, then remove all subscriptions from the client
	// else loop over, both the client's subscription list and given array of channels and make a new list, and assign it
	if len(channels) == 0 {
		// no active subscriptions left --> turning OFF the subscribedMode
		client.Subscriptions = []string{}
		channels = append([]string(nil), client.Subscriptions...)
		client.inSubscription = false
	} else {
		newSubscriptionList := []string{}

		// no duplication, no nested loop, O(N) complexity
		toRemove := make(map[string]bool)
		for _, chnl := range channels {
			toRemove[chnl] = true
		}

		for _, sub := range client.Subscriptions {
			if !toRemove[sub] {
				newSubscriptionList = append(newSubscriptionList, sub)
			}
		}
		client.Subscriptions = newSubscriptionList
	}

	if len(client.Subscriptions) == 0 {
		client.inSubscription = false
	}

	// removing the clients from channels in the global registry
	for _, channel := range channels {
		clients := r.pubsubs.channels[channel]
		newList := []*Client{}
		for _, clnt := range clients {
			if clnt != client {
				newList = append(newList, clnt)
			}
		}

		// if there are no active clients left in the channel, it is deleted
		if len(newList) == 0 {
			delete(r.pubsubs.channels, channel)
		} else {
			r.pubsubs.channels[channel] = newList
		}
	}

	remaining := len(client.Subscriptions)
	var response string
	for _, chnl := range channels {
		response += fmt.Sprintf("*3\r\n$11\r\nunsubscribe\r\n$%d\r\n%s\r\n:%d\r\n", len(chnl), chnl, remaining)
	}

	return response, true
}
