# Cluster Map

Cluster map contains current cluster information, which includes node and group.
It is maintained by per node: mds, ds and gw.

## Use cases

```
// Get the latest cluster map.
clusterMap = cmap.GetLatest()

// Get the latest cluster map from the given mds address.
clusterMap = cmap.GetLatest(mdsAddr)

// Get the human readable string of cluster map.
fmt.Println(clusterMap.HumanReadable())

// Create search call.
searchCall := clusterMap.SearchCall()

// Search call will find mds type and alive status of node.
searchCall.Type(clusterMap.MDS).Status(clusterMap.Alive)

// Get the node that matches a search call's conditions.
node := searchCall.Do()
```
