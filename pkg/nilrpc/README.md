# RPC naming convention

## Prefixes

This is a list of domain handlers and method prefixes.

| Domain         | Method prefix  | Acronym | Description                                                                 |
| -------------- | -------------- | ------- | --------------------------------------------------------------------------- |
| MDS cluster    | MDS_CLUSTER    | MCL     | The mds cluster handler is the collection of routines for handling cluster management related requests. |
| MDS account    | MDS_ACCOUNT    | MAC     | The mds user handler is the collection of routines for handling user related requests. |
| MDS object     | MDS_OBJECT     | MOB     | The mds object handler is the collection of routines for handling object related requests. |
| MDS encoding   | MDS_ENCODING   | MEN     | The mds encoding handler is the collection of routines for handling global encoding requests. |
| DS cluster     | DS_CLUSTER     | DCL     | The ds cluster handler is the collection of routines for handling cluster management related requests. |

## Methods

"Method prefix" + "." + "Verb and noun"

Use predefined method prefixes and write clear verb and noun by camel case style.

### Examples 

* Get cluster map from the MDS
  * MDS_CLUSTER.GetClusterMap
* Do rebalance from the MDS
  * MDS_CLUSTER.DoRebalance

## Requests

"Acronym of method prefix" + "Verb and noun" + "Request or Response"

Use predefined acronym of method prefixes and write clear verb and noun by camel case style.

### Examples

* MCLGetClusterMapRequest
* MCLGetClusterMapResponse
* DCLAddVolumeRequest
* DCLAddVolumeResponse
