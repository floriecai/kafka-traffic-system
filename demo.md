Demo Plan: 

1) Change config.json to `cluster-size`: 8, `min-replicas`: 4
2) Start nodes 1- 8
3) Run masterTestApp.go (NEED TO COMMENT OUT THE WRITING TO INTERNAL CONN)
4) Run consumerApp.go

Should see the map filled with points

Failures: 

1) Refresh browser page - should be empty now (because we haven't re read the data and populated it)
2) Kill 1 follower
3) Kill Leader
3) Run consumerApp.go 
    Show that can still read data even though Leader + Follower died.

4) Kill Server
    Show that still operational with server dead (need to refresh browser to read again)

Rejoining Nodes:
5) Server comes back online
   Operational again

6) One of the old Followers comes back online 
   Back to being a part of the cluster

7) Make another write, show that Leaders + all Followers (including rejoined one) 
   received the writes