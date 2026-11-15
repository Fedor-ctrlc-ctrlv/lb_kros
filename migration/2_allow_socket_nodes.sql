BEGIN;

ALTER TABLE fs_nodes DROP CONSTRAINT fs_nodes_node_type_check;

ALTER TABLE fs_nodes ADD CONSTRAINT fs_nodes_node_type_check
    CHECK (node_type IN ('file', 'directory', 'symlink', 'socket'));

COMMIT;