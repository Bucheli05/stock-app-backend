#!/bin/bash

# Wait for CockroachDB to be ready
sleep 10

# Create the database if it doesn't exist
cockroach sql --insecure --host=localhost:26257 << EOF
CREATE DATABASE IF NOT EXISTS my_new_db;
EOF

echo "Database initialization complete"
