## Bulk Data Insertion Benchmark for PostgreSQL

### Overview
A high-performance benchmarking tool designed to evaluate and compare different bulk insertion methods in PostgreSQL. The tool provides empirical data to help determine the optimal insertion strategy for large-scale data loading scenarios.

### Key Features

#### Insertion Methods
| Method        | Description                                                                 | Best For                  |
|---------------|-----------------------------------------------------------------------------|---------------------------|
| `copyfrom`    | Direct PostgreSQL COPY protocol (most efficient for raw speed)              | Large, simple data loads  |
| `pgxbatch`    | Batched prepared statements using pgx library                               | Medium-sized transactions |
| `unnestbatch` | Array-based bulk operations using UNNEST                                    | Complex data structures   |

#### Benchmarking Capabilities
- Configurable batch sizes (100-50,000 records)
- Memory and CPU profiling integration
- Pipeline mode for concurrent processing
- Clean environment management (`--truncate`)

### Installation & Setup

```bash
# Clone and build
git clone https://github.com/aaa2ppp/pg-bulk-flow.git
cd pg-bulk-flow
make build

# Configure environment
cp env.example .env
nano .env  # Set your DB parameters

# Start test environment
make migration-up  # Launches PostgreSQL in Docker
```

### Usage Examples

#### Basic Benchmark
```bash
./bin/fillnames -method copyfrom -batch 5000 -truncate
```

#### Comparative Analysis
```bash
# Test all methods with 10k batches
for method in copyfrom pgxbatch unnestbatch; do
  ./bin/fillnames -method $method -batch 10000 -truncate | tee results_${method}.json
done
```

#### Advanced Profiling
```bash
# CPU profiling
./bin/fillnames -method pgxbatch -cpuprofile=pgx_cpu.pprof

# Memory analysis
./bin/fillnames -method unnestbatch -memprofile=unnest_mem.pprof
```

### Performance Metrics
The tool outputs detailed statistics including:
- Total execution time
- Records inserted per second
- Memory allocation
- Batch processing times
- Pipeline efficiency (when enabled)

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

### Interpretation Guide
Typical performance characteristics:

1. **Small Batches (<1k records)**
   - `pgxbatch` often performs best
   - Low memory overhead

2. **Medium Batches (1k-10k records)**
   - `copyfrom` begins to dominate
   - `unnestbatch` shows consistent performance

3. **Large Batches (>10k records)**
   - `copyfrom` is typically fastest
   - Memory usage becomes significant factor

### Visualization
For results analysis, consider:

```bash
# Generate comparative charts
go tool pprof -png -output=profile.png *.pprof

# Process JSON results
jq '{method: .config.method, records_sec: (.stats.inserted/(.stats.elapsed/1000))}' *.json
