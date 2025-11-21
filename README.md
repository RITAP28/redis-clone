# 🧠 Redis Clone in Go  
*A high-performance, educational Redis implementation built from scratch in Golang.*

---

## 🚀 What is This Project?

This is a **Redis clone written entirely in Go** — built to understand how Redis works under the hood.

It includes:

- TCP server  
- RESP protocol parsing  
- In-memory key-value store  
- Data structures (Strings, Lists, Sets, Hashes)  
- Expiry system + background deletion  
- JSON-based RDB persistence  
- Pub/Sub  
- Transactions (MULTI / EXEC)  

And yes — you can use the **official `redis-cli`** to talk to this server.

---

## 📊 Feature Comparison

| Feature | Real Redis | My Clone |
|--------|------------|----------|
| In-memory storage | ✅ | ✅ |
| Strings | ✅ | ✅ |
| Lists | ✅ | ✅ |
| Sets | ✅ | ✅ |
| Hashes | ✅ | ✅ |
| Sorted Sets | ✅ | ⏳ planned |
| Expiry (EXPIRE/TTL) | ✅ | ✅ |
| Persistence (RDB/AOF) | RDB + AOF | JSON RDB (AOF planned) |
| Pub/Sub | ✅ | ✅ |
| Transactions | ✅ | ✅ |
| Streams | ✅ | ⏳ planned |
| Replication | ✅ | ❌ |
| Lua Scripting | ✅ | ❌ |
| AUTH / ACL | ✅ | ❌ |
| Cluster Mode | ✅ | ❌ |

> 🟢 **Legend:**  
> `✅ implemented` `⏳ planned` `❌ not implemented`

---

## 🧩 Supported Data Structures

- 📕 **Strings**  
- 📚 **Lists**  
- 🧺 **Sets**  
- 🗂️ **Hashes**

(Sorted Sets coming soon!)

---

# 🧠 Supported Commands (Full List)

Below is a complete list of all commands implemented in this Redis clone.

---

<details>
<summary><strong>🔴 Core Commands</strong></summary>

| Command | Description |
|--------|-------------|
| `PING` | Connectivity test |
| `SET key value` | Set a key |
| `GET key` | Get a key |
| `DELETE key` | Delete a key |
| `EXPIRE key seconds` | Set TTL in seconds |
| `PEXPIRE key ms` | Set TTL in ms |
| `TTL key` | Time-to-live (seconds) |
| `PTTL key` | Time-to-live (ms) |

</details>

---

<details>
<summary><strong>🟩 List Commands</strong></summary>

| Command | Description |
|--------|-------------|
| `LPUSH key value [value ...]` | Push values to left |
| `RPUSH key value [value ...]` | Push values to right |
| `LPOP key` | Remove and return first element |
| `RPOP key` | Remove and return last element |
| `LRANGE key start stop` | Read list slice |
| `LLEN key` | List length |

</details>

---

<details>
<summary><strong>🟦 Set Commands</strong></summary>

| Command | Description |
|--------|-------------|
| `SADD key member [member ...]` | Add members to set |
| `SREM key member [member ...]` | Remove members |
| `SMEMBERS key` | Get all members |
| `SISMEMBER key member` | Check membership |
| `SCARD key` | Set cardinality |

</details>

---

<details>
<summary><strong>🟨 Hash Commands</strong></summary>

| Command | Description |
|--------|-------------|
| `HSET key field value [field value ...]` | Set hash fields |
| `HGET key field` | Get a specific field |
| `HGETALL key` | Get all fields |
| `HDEL key field [field ...]` | Delete fields |
| `HLEN key` | Number of fields |

</details>

---

<details>
<summary><strong>🟧 Transaction Commands</strong></summary>

| Command | Description |
|--------|-------------|
| `MULTI` | Begin transaction |
| `EXEC` | Execute all queued commands |
| `DISCARD` | Cancel transaction |
| `INCR key` | Increment integer value |

All valid commands can be queued between `MULTI` and `EXEC`.

</details>

---

<details>
<summary><strong>🟪 Pub/Sub Commands</strong></summary>

| Command | Description |
|--------|-------------|
| `SUBSCRIBE channel [channel ...]` | Subscribe to channels |
| `UNSUBSCRIBE [channel ...]` | Unsubscribe (or all) |
| `PUBLISH channel message` | Broadcast a message |

</details>

---

## 💾 Persistence

This clone supports **RDB-like persistence** using readable JSON files.

Snapshot example (`dump.rdb.json`):

```json
{
  "numbers": {
    "value": ["one", "two", "three"],
    "expiryTime": "0001-01-01T00:00:00Z"
  },
  "users:registered": {
    "value": {
      "user:alpha": {},
      "user:beta": {},
      "user:gamma": {}
    },
    "expiryTime": "0001-01-01T00:00:00Z"
  }
}

```

## Expiry & Background Cleaner

A background goroutine periodically removes expired keys:

```go

redisServer.StartExpiryCleaner(20 * time.Second)

```

## Connecting with redis-cli

Run the server:

```bash

go run main.go

```

Use the official Redis CLI:

```bash

redis-cli -p 8080

```

Example:

```text

127.0.0.1:8080> SET hello world
OK
127.0.0.1:8080> GET hello
"world"
127.0.0.1:8080> EXPIRE hello 5
(integer) 1

```

## Architecture Overview:

```txt

Client (redis-cli)
        |
        v
+---------------------------+
|    Redis Clone Server     |
+---------------------------+
| RESP Protocol Parser      |
| Command Dispatcher        |
| In-Memory Map Store       |
| Expiry Engine             |
| JSON RDB Persistence      |
| Transactions Engine       |
| Pub/Sub Subsystem         |
+---------------------------+

```

## 🤝 Contributing

This is a learning project — contributions, suggestions, and improvements are welcome.

Fork the repo → open a PR!
Clone the repo → advance it further!

## 🧾 License

License under the MIT License.

## ✨ Final Note

The best way to understand a system is to build it from scratch — even if it breaks a hundred times along the way.