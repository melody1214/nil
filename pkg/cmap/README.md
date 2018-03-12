# Cluster Map

Cluster map contains current cluster information, which includes node and group.
It is maintained by per node: mds, ds and gw.

## Use cases

```
// Create new cluster map with the known mds address.
clusterMap := cmap.New(mdsAddr)

// Get the latest cluster map.
clusterMap.GetLatest()

// Store cluster map into the disk.
clusterMap.Store()

// Create search call.
searchCall := clusterMap.SearchCall()

// Search call will find mds type and alive status of node.
searchCall.Type(clusterMap.MDS).Status(clusterMap.Alive)

// Get the node that matches a search call's conditions.
node := searchCall.Do()
```
