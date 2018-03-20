# Cluster Map

Cluster map contains current cluster information, which includes node and group.
It is maintained by per node: mds, ds and gw.

## Use cases

```
// Get the latest cluster map from the remote.
clusterMap = cmap.GetLatest(FromRemote())

// Get the latest cluster map from the remote with the given mds address.
clusterMap = cmap.GetLatest(FromRemote(mdsAddr))

// Get the latest cluster map from the local file.
// Note: default option is from local.
clusterMap = cmap.GetLatest(FromLocal())

// Get the human readable string of cluster map.
fmt.Println(clusterMap.HumanReadable())

// Create search call.
searchCall := clusterMap.SearchCall()

// Search call will find mds type and alive status of node.
searchCall.Type(clusterMap.MDS).Status(clusterMap.Alive)

// Get the node that matches a search call's conditions.
node, err := searchCall.Do()
if err != nil {
    // Handling error.
}
```
