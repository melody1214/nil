# Membership Management

All cluster members are manged by membership protocol and write their all information in cluster map (cmap).

## Cluster Map

Cluster map includes information about nodes, volume, and encoding groups. 

### Version

Cluster map version is increased monotonically when the master node makes three kinds of change.

* Add new member (node, volume, encoding group)
* Delete member (node, volume, encoding group)
* Change volumes of the encoding group

Each version of cluster maps are created, they are stored in the filesystem and propagated via gossiping.

### Update

Each member can update some self related information without update version (without intervention of the master node).

* Encoding group leader can update its capacity information
* Each node can update the status of another node they discover
