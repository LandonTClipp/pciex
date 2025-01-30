package main

// These imports will be used later on the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.
import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/LandonTClipp/pciex/models"
	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v2"
)

type Slot struct {
	Domain string
	Bus    string
	Device string
	Func   string
}

func NewSlotFromString(s string) Slot {
	split := strings.Split(s, ":")
	deviceFunc := split[2]
	deviceFuncSplit := strings.Split(deviceFunc, ".")

	return Slot{
		Domain: split[0],
		Bus:    split[1],
		Device: deviceFuncSplit[0],
		Func:   deviceFuncSplit[1],
	}
}

type PCIElement struct {
	Slot   Slot
	Class  string
	Vendor string
	Device string
}

type Details struct {
	Id            string
	Class         string
	Claimed       bool
	Handle        string
	Description   string
	Product       string
	Vendor        string
	Physid        string
	Businfo       string
	Version       string
	Width         int
	Clock         int
	Serial        string
	Slot          string
	Units         string
	Size          int
	Configuration map[string]any
	Capabilities  map[string]any
}

func (d Details) String() string {
	var s string
	s += d.Class + " | "
	switch d.Class {
	case "bridge":
		s += d.Handle
	case "bus":
		s += d.Description
	case "display", "memory", "communication", "generic", "network":
		s += d.Product
	default:
		s += d.Description
	}
	return s
}

type LshwElem struct {
	Details
	Children []LshwElem
}

func buildPCIETreeHelper(parent *models.Node, children []LshwElem) error {
	if parent == nil {
		panic("parent is nil")
	}
	for _, child := range children {
		out, err := yaml.Marshal(child.Details)
		if err != nil {
			return fmt.Errorf("unmarshalling yaml: %w", err)
		}
		childNode := parent.AddChild(child.Details.String(), string(out))
		if err := buildPCIETreeHelper(childNode, child.Children); err != nil {
			return err
		}
	}
	return nil
}

func buildPCIETreeV2(tree *models.TreeModel) error {
	cmd := exec.Command("/usr/bin/lshw", "-json")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("reading command output: %w", err)
	}
	lshw := LshwElem{}
	if err := json.Unmarshal(out, &lshw); err != nil {
		// Some versions of lshw wrap the JSON in a single-element array.
		// Why? I don't know. See if we can successfully unmarshal into an
		// array.
		lshwArray := []LshwElem{}
		if newErr := json.Unmarshal(out, &lshwArray); newErr != nil {
			return fmt.Errorf("unmarshalling json: %w", err)
		}
		lshw = lshwArray[0]
	}
	tree.Root = models.NewNode("root", "", nil, tree)
	// Find PCI element
	for _, child := range lshw.Children[0].Children {
		if !strings.HasPrefix(child.Id, "pci") {
			continue
		}
		out, err := yaml.Marshal(child.Details)
		if err != nil {
			return fmt.Errorf("unmarshalling yaml: %w", err)
		}
		childNode := tree.Root.AddChild(child.String(), string(out))
		if err := buildPCIETreeHelper(childNode, child.Children); err != nil {
			return err
		}
	}
	return nil
}

func buildPCIETree(tree *models.TreeModel) error {
	cmd := exec.Command("/usr/bin/lspci", "-mm", "-nn", "-q", "-D", "-PP")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("reading command output: %w", err)
	}

	reader := csv.NewReader(bytes.NewReader(out))
	reader.Comma = ' '
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println(string(out))
		return fmt.Errorf("reading CSV: %w", err)
	}
	for _, record := range records {
		element := PCIElement{
			Slot:   NewSlotFromString(record[0]),
			Class:  record[1],
			Vendor: record[2],
			Device: record[3],
		}
		yamlRepr, err := yaml.Marshal(element)
		if err != nil {
			return fmt.Errorf("marshalling yaml: %w", err)
		}
		tree.Root.AddChild(element.Class, string(yamlRepr))
	}
	return nil
}

func main() {
	rootModel := models.NewRootModel()
	root := models.NewNode("Root", "", nil, rootModel.Tree)
	rootModel.Tree.Root = root

	//if err := buildPCIETree(rootModel.Tree); err != nil {
	//	fmt.Printf("Error occurred: %v\n", err)
	//	os.Exit(1)
	//}
	if err := buildPCIETreeV2(rootModel.Tree); err != nil {
		fmt.Printf("Error occurred: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(rootModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error occurred: %v\n", err)
		os.Exit(1)
	}
}
