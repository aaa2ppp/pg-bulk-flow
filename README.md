## Bulk Data Insertion Benchmark for PostgreSQL

### Overview
A high-performance benchmarking tool designed to evaluate and compare different bulk insertion methods in PostgreSQL.
The tool provides empirical data to help determine the optimal insertion strategy for large-scale data loading scenarios.

### Key Features

#### Insertion Methods
| Method        | Description                                    |
|---------------|------------------------------------------------|
| `copyfrom`    | Direct PostgreSQL COPY protocol                |
| `pgxbatch`    | Batched prepared statements using pgx library  |
| `unnestbatch` | Array-based bulk operations using UNNEST       |

#### Benchmarking Capabilities
- Stream processing
- Configurable batch sizes (for pgxbatch and unnestbatch)
- Memory and CPU profiling integration
- Pipeline mode for concurrent processing
- Clean environment management (`--truncate`)

#### Performance Metrics
The tool outputs detailed statistics including:
- Total execution time
- Records inserted
- CPU usage
- Memory allocation

### Installation & Setup

```bash
# Clone and build
git clone https://github.com/aaa2ppp/pg-bulk-flow.git
cd pg-bulk-flow
make build

# Configure environment
cp env.example .env
nano .env  # Set your DB parameters and others

# Start test environment
make db-up # Launches PostgreSQL in Docker
make migrate-up

# or (if an external database is used)
make migrate-up USE_EXTERNAL_DB=yes # will be used DB_ADDR from .env
```

### Usage Examples

#### Basic Benchmark
```bash
./bin/fillnames -method pgxbatch -batch 5000 -truncate
```

#### Comparative Analysis
```bash
mkdir -p ./tmp

# Test all methods with 10k batches
for method in copyfrom pgxbatch unnestbatch; do
  ./bin/fillnames -method $method -batch 10000 -truncate | tee ./tmp/results_${method}.json
done
```

#### Advanced Profiling
```bash
mkdir -p ./tmp

# CPU profiling
./bin/fillnames -method pgxbatch -cpuprofile=./tmp/pgx_cpu.pprof

# Memory analysis
./bin/fillnames -method unnestbatch -memprofile=./tmp/unnest_mem.pprof
```

#### Visualization
For results analysis, consider:

```bash
# Generate comparative charts
find ./tmp -name '*.pprof' | while read -r file; do
  go tool pprof -png -output="${file%.pprof}.png" ./bin/fillnames "$file"
done

# Process JSON results
jq '{method: .config.method, records_sec: (.stats.inserted/(.stats.elapsed/1000))}' ./tmp/*.json
```

### Technical Considerations

#### Test Environment
- Dedicated test table (`names`)
- Constraints intentionally disabled
- Simple schema for focused benchmarking
- Dockerized PostgreSQL for consistency

#### Limitations
- Not designed for production data loading
- Doesn't handle duplicates/constraints
- Optimized for insertion speed comparison only
