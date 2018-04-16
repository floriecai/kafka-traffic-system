Demo Plan: 


1) Start nodes 1- 8
2) Run masterTestApp.go
3) Run consumerApp.go

Should see the map filled with points

4) Write one point to show it works still

Failures: 

1) Refresh browser page - should be empty now (because we haven't re read the data and populated it)
2) Kill 1 follower
3) Write one point
5) Kill Leader
6) Write one point
7) Kill Server
8) Write one point

Rejoining Nodes:
1) Server comes back online
2) write one point
3) Follower comes backo nline
4) Write one point