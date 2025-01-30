# pciex

PCI Explorer is a PCIe topology explorer and visualizer. The generated tree can be visualized and each element inspected using a simple terminal UI designed with the bubbletea framework.

<img width="1500" alt="Screenshot 2025-01-29 at 6 57 09 PM" src="https://github.com/user-attachments/assets/e055812c-c6f4-4a19-912c-49eebc997802" />

## Installation

This can be installed using:

```
GOPROXY=direct go run github.com/LandonTClipp/pciex@latest
```

## Dependencies

This module requires the `lshw` utility to be installed.

### Ubuntu

```
sudo apt-get install lshw
```

### Redhat

```
sudo yum install lshw
```
