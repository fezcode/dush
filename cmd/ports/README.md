# Ports Utility

A cross-platform utility to list open ports and their associated processes.

## Usage

```bash
ports [options]
```

### Options:
- `-tcp`: Show TCP ports (default: true)
- `-udp`: Show UDP ports (default: true)
- `-listen`: Show only listening ports (default: false)
- `-pid`: Show PID and process information (default: true)

## Platform Support
- **Windows**: Uses native `iphlpapi.dll` (GetExtendedTcpTable/GetExtendedUdpTable)
- **Linux**: Parses `/proc/net/tcp`, `/proc/net/udp` and maps to PIDs via `/proc/[pid]/fd`
- **macOS**: Uses native `proc_info` syscall (SYS_PROC_INFO)
