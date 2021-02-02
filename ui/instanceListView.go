package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type InstanceListView struct {
	*tview.Grid
	instances      []*ec2.Instance
	currentItemIdx int
	items          []tview.Primitive
}

func (v *InstanceListView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handler := v.Grid.InputHandler()

	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyTAB:
			idx := (v.currentItemIdx + 1) % len(v.items)
			v.currentItemIdx = idx
			setFocus(v.items[idx])
		}

		handler(event, setFocus)
	}
}

func (v *InstanceListView) Focus(delegate func(p tview.Primitive)) {
	delegate(v.items[v.currentItemIdx])
}

func NewInstanceListView(ui *Ui, instances []*ec2.Instance) *InstanceListView {
	var filteredInstances *[]*ec2.Instance
	filteredInstances = &instances

	table := tview.NewTable().
		SetSelectable(true, false)

	table.SetBorder(true)
	table.SetFixed(1, 1)
	table.SetSelectedFunc(selected(ui, filteredInstances))

	displayInstances(table, instances)

	filterInput := tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetLabel("Filter").
		SetLabelWidth(8)

	filterInput.SetChangedFunc(filterChangedHander(table, instances, filteredInstances))

	grid := tview.NewGrid()

	grid.
		SetBorders(true).
		SetRows(1, -1, 1).
		SetColumns(1, -1, -1, 1)

	grid.AddItem(filterInput, 0, 1, 1, 2, 0, 0, false)
	grid.AddItem(table, 1, 1, 1, 2, 0, 0, false)

	items := []tview.Primitive{filterInput, table}

	return &InstanceListView{
		grid,
		instances,
		0,
		items,
	}
}

func selected(ui *Ui, instances *[]*ec2.Instance) func(row, coln int) {
	return func(row, col int) {
		instance := (*instances)[row-1]
		instanceId := aws.StringValue(instance.InstanceId)

		modal := tview.NewModal().
			SetText(fmt.Sprintf("Do you connecto to '%s'?", instanceId)).
			AddButtons([]string{"Yes", "Cancel"}).
			SetDoneFunc(func(index int, label string) {
				ui.removeModal()

				if label == "Yes" {
					ui.app.Suspend(func() {
						command := fmt.Sprintf("aws ssm start-session --target %s", instanceId)
						cmd := exec.Command("bash", "-c", command)
						cmd.Stdin = os.Stdin
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						cmd.Run()
					})
				}
			})

		ui.showModal(modal)
	}
}

func displayInstances(table *tview.Table, instances []*ec2.Instance) {
	table.Clear()

	table.SetCell(0, 0, tview.NewTableCell("Name").SetSelectable(false))
	table.SetCell(0, 1, tview.NewTableCell("instance-id").SetSelectable(false))

	for i, instance := range instances {
		addInstanceToTable(table, i+1, instance)
	}
}

func addInstanceToTable(table *tview.Table, row int, instance *ec2.Instance) {
	table.SetCell(row, 0, tview.NewTableCell(describeInstanceName(instance)))
	table.SetCell(row, 1, tview.NewTableCell(aws.StringValue(instance.InstanceId)))
}

func describeInstanceName(instance *ec2.Instance) string {
	for _, tag := range instance.Tags {
		if aws.StringValue(tag.Key) == "Name" {
			return aws.StringValue(tag.Value)
		}
	}
	return ""
}

func doneHandler(input *tview.InputField, table *tview.Table, instances *[]string) func(key tcell.Key) {
	return func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		text := input.GetText()

		*instances = append(*instances, text)

		table.SetCellSimple(table.GetRowCount()+1, 0, text)
		input.SetText("")
	}
}

func filterChangedHander(table *tview.Table, instances []*ec2.Instance, filteredInstances *[]*ec2.Instance) func(text string) {
	return func(text string) {
		*filteredInstances = nil

		if text == "" {
			*filteredInstances = instances
			displayInstances(table, instances)
			return
		}

		for _, instance := range instances {
			if strings.Index(describeInstanceName(instance), text) != -1 {
				*filteredInstances = append(*filteredInstances, instance)
			}
		}

		displayInstances(table, *filteredInstances)

		return
	}
}
