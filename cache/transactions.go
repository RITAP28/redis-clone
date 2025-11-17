package cache

// in transactions, commands are queued and executed sequentially atomically, basically all commands are executed at once or none at all
// MULTI --> starts a transaction, turns on the inTransaction mode and starts queueing
// EXEC --> after all the commands are queued, they are executed sequentially and returns all the results
// DISCARD --> Clear queue, exits transaction mode

func(r *RedisCache) queueCommands(client *Client, cmd []any) bool {
	client.Transactions = append(client.Transactions, cmd)
	return true
}

func(r *RedisCache) MULTI(client *Client) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// since MULTI calls cannot be nested
	if client.InTransaction {
		return "", false
	}

	// set the transaction mode to true
	// transaction queue for commands is initialised
	client.InTransaction = true
	client.Transactions = make([][]interface{}, 0)
	return "MULTI command is active", true
}















