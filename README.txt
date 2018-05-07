Kafka Traffic System

There are 2 client types: Producers & Consumers. Producers write to the data-stream (Topic), Consumers read from the stream. Based on the config settings, writes are replicated and can survive up to N failures in a cluster, where the cluster size is 2N + 1.


// Requirements for config.json 

```
{
    "rpc-ip-port": ":12345",
    "node-settings": {
        "heartbeat": 10000,
        "min-replicas": 3,
        "cluster-size": 5
    }
}
```

`cluster-size` is the number of nodes in a cluster *including* the leader. 
So if `cluster-size` is 3, there will be 1 Leader node and 2 Follower nodes.

`min-replicas` is the minimum number of nodes a Write must be replicated
on in order for a Write call to succeed. For our system, we assume
the number is at least a majority of Follower nodes. i.e. If `cluster-size`
is 5, there are 4 Follower nodes. `min-replicas` >= 3.
We allow users to set a number higher in order to guarantee better replication
for Writes. 
