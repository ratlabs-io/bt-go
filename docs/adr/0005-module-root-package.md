# Library package at module root

Sources live at the repo root as `package bt` with import path `github.com/ratlabs-io/bt-go` (not `.../bt-go/bt`). Chosen for simpler `go get` / import ergonomics while the surface is still small. Revisit only if the root becomes crowded with non-library files.
