# gocopy

Parallel directory copy written in Go.

On Linux, uses io.Copy and uses CopyFileW kernel32.dll function on windows.