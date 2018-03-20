# Cluster Map

Cluster map contains current cluster information, which includes node and group.
It is maintained by per node: mds, ds and gw.

## Use cases

```
// Get the initial cluster map with the given coordinator address.
err := cmap.Initial(coordinator)
if err != nil {
    // Handling error
}

// Get the latest cluster map from the remote.
// Note: default option is from local.
clusterMap := cmap.GetLatest(WithFromRemote(true))

// Get the latest cluster map from the local file.
clusterMap = cmap.GetLatest()

// Get the human readable string of cluster map.
fmt.Println(clusterMap.HumanReadable())

// Create search call.
searchCall := clusterMap.SearchCall()

// Search call will find mds type and alive status of node.
searchCall.Type(clusterMap.MDS).Status(clusterMap.Alive)

// Get the node that matches a search call's conditions.
node, err := searchCall.Do()
if err != nil {
    // Handling error
}

// Save the cluster map to the local file system.
err := clusterMap.Save()
if err != nil {
    // Handling error
}
```
