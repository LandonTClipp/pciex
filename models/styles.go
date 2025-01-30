package models

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/chigopher/pathlib"
)

var (
	itemStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	itemStyleSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Bold(true)
	viewportStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).
			Border(lipgloss.NormalBorder())
	enumeratorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).MarginRight(1)
	rootStyle       = itemStyle
)

var nvidiaSMIExample string = `
Thu May 10 09:05:07 2018
+-----------------------------------------------------------------------------+
| NVIDIA-SMI 384.111                Driver Version: 384.111                   |
|-------------------------------+----------------------+----------------------+
| GPU  Name        Persistence-M| Bus-Id        Disp.A | Volatile Uncorr. ECC |
| Fan  Temp  Perf  Pwr:Usage/Cap|         Memory-Usage | GPU-Util  Compute M. |
|===============================+======================+======================|
|   0  GeForce GTX 108...  Off  | 00000000:0A:00.0 Off |                  N/A |
| 61%   74C    P2   195W / 250W |   5409MiB / 11172MiB |    100%      Default |
+-------------------------------+----------------------+----------------------+

+-----------------------------------------------------------------------------+
| Processes:                                                       GPU Memory |
|  GPU       PID   Type   Process name                             Usage      |
|=============================================================================|
|    0      5973      C   ...master_JPG/build/tools/program_pytho.bin  4862MiB |
|    0     46324      C   python                                       537MiB |
+-----------------------------------------------------------------------------+
`

func itemStyleFunc(children tree.Children, i int) lipgloss.Style {
	child := children.At(i)
	if n, ok := child.(*Node); ok {
		debug(fmt.Sprintf("n.model.CurNode: %p \tn: %p\n", n.model.CurNode, n))
		if n.model.CurNode == n {
			return itemStyleSelected
		}
	} else {
		debug(fmt.Sprintf("Is not *Node. Is: %T\n", child))
	}

	return itemStyle
}

func debug(s string) {
	return

	out := pathlib.NewPath("out.txt")
	file, err := out.OpenFile(os.O_APPEND | os.O_WRONLY | os.O_CREATE)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	if _, err := file.WriteString(s); err != nil {
		panic(err)
	}
}
