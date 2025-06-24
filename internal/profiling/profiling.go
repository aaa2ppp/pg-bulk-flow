package profiling

import (
	"bufio"
	"flag"
	"log/slog"
	"os"
	"runtime"
	"runtime/pprof"
)

var (
	cpuprofile   = flag.String("cpuprofile", "", "write CPU profile to `file`")
	memprofile   = flag.String("memprofile", "", "write memory profile to `file`")
	blockprofile = flag.String("blockprofile", "", "write block profile to `file`")
)

// Start starts profiling if the -cpuprofile, -memprofile or -blockprofile flags are set
// To use it, add the following code to your main function:
//
//	flags.Parse()
//	stopProfiling := profiling.Start()
//	defer stopProfiling()
//	// you code... Do not use os.Exit or log.Fatal. Because defer doesn't intercept them.
func Start() func() {
	stopCPUProfile := startCPUProfile()
	stopBlockProfile := startBlockProfile()
	return func() {
		if stopBlockProfile != nil {
			stopBlockProfile()
		}
		if stopCPUProfile != nil {
			stopCPUProfile()
		}
		saveMemoryProfile()
	}
}

// Do simple wrap you code (see Start)
// Example:
//
//	flags.Parse()
//	profiling.Do(func() {
//		// you code...
//	})
func Do(fn func()) {
	stop := Start()
	defer stop()
	fn()
}

func startCPUProfile() func() {
	if *cpuprofile == "" {
		return nil
	}

	f, err := os.Create(*cpuprofile)
	if err != nil {
		slog.Error("could not create CPU profile", "error", err)
		return nil
	}
	r := bufio.NewWriter(f)

	if err := pprof.StartCPUProfile(f); err != nil {
		slog.Error("could not start CPU profile", "error", err)
		f.Close()
		return nil
	}

	return func() {
		pprof.StopCPUProfile()
		if err := r.Flush(); err != nil {
			slog.Error("could not flush CPU profile", "error", err)
		}
		if err := f.Close(); err != nil {
			slog.Warn("could not close CPU profile", "error", err)
		}
	}
}

func saveMemoryProfile() {
	if *memprofile == "" {
		return
	}

	f, err := os.Create(*memprofile)
	if err != nil {
		slog.Error("could not create memory profile", "error", err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			slog.Warn("could not close memory profile", "error", err)
		}
	}()

	runtime.GC()
	if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
		slog.Error("could not write memory profile", "error", err)
		return
	}
}

func startBlockProfile() func() {
	if *blockprofile == "" {
		return nil
	}
	runtime.SetBlockProfileRate(1)
	return func() {
		runtime.SetBlockProfileRate(0)
		saveBlockProfile()
	}
}

func saveBlockProfile() {
	f, err := os.Create(*blockprofile)
	if err != nil {
		slog.Error("could not create block profile", "error", err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			slog.Warn("could not close block profile", "error", err)
		}
	}()

	if err := pprof.Lookup("block").WriteTo(f, 0); err != nil {
		slog.Error("could not write block profile", "error", err)
		return
	}
}
