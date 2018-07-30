package mysql

import (
	"github.com/chanyoung/nil/app/mds/application/cluster"
)

type clusterStore struct {
	*Store
}

// NewClusterRepository returns a new instance of a mysql cluster map repository.
func NewClusterRepository(s *Store) cluster.Repository {
	return &clusterStore{
		Store: s,
	}
}

// func (s *clusterStore) FindAllNodes(txid repository.TxID) (nodes []cmap.Node, err error) {
// 	q := fmt.Sprintf(
// 		`
// 		SELECT
// 			node_id,
// 			node_name,
//             node_type,
// 			node_status,
// 			node_address
// 		FROM
// 			node
// 		`,
// 	)

// 	var rows *sql.Rows
// 	rows, err = s.Query(txid, q)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	nodes = make([]cmap.Node, 0)
// 	for rows.Next() {
// 		n := cmap.Node{}

// 		if err = rows.Scan(&n.ID, &n.Name, &n.Type, &n.Stat, &n.Addr); err != nil {
// 			return nil, err
// 		}

// 		nodes = append(nodes, n)
// 	}

// 	return
// }

// func (s *clusterStore) GetNewClusterMapVer(txid repository.TxID) (cmap.Version, error) {
// 	q := fmt.Sprintf(
// 		`
// 		INSERT INTO cmap (cmap_id)
// 		VALUES (NULL)
// 		`,
// 	)

// 	res, err := s.Execute(txid, q)
// 	if err != nil {
// 		return -1, err
// 	}

// 	ver, err := res.LastInsertId()
// 	if err != nil {
// 		return -1, err
// 	}

// 	return cmap.Version(ver), nil
// }

// func (s *clusterStore) LocalJoin(node cmap.Node) error {
// 	q := fmt.Sprintf(
// 		`
// 		INSERT INTO node (node_name, node_type, node_status, node_address)
// 		VALUES ('%s', '%s', '%s', '%s')
// 		`, node.Name, node.Type, node.Stat, node.Addr,
// 	)

// 	_, err := s.Execute(repository.NotTx, q)
// 	return err
// }

// func (s *clusterStore) GlobalJoin(raftAddr, nodeID string) error {
// 	return s.Store.Join(nodeID, raftAddr)
// }
