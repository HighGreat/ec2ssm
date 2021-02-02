package ui

import (
	"ec2ssm/aws"

	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
)

type Ui struct {
	app      *tview.Application
	rootPage *tview.Pages
	pages    []*tview.Primitive
	ec2svc   *aws.Ec2Svc
}

func (u *Ui) Run() error {
	runewidth.DefaultCondition.EastAsianWidth = false

	if err := u.init(); err != nil {
		return err
	}

	if err := u.app.SetRoot(u.rootPage, true).Run(); err != nil {
		u.app.Stop()
		return err
	}

	return nil
}

func (u *Ui) init() error {
	instances, err := u.ec2svc.FetchInstances()
	if err != nil {
		return err
	}

	instanceListView := NewInstanceListView(u, instances)

	u.rootPage.AddAndSwitchToPage("instance-list-view", instanceListView, true)

	return nil
}

func (u *Ui) showModal(p tview.Primitive) {
	u.rootPage.AddAndSwitchToPage("modal", modal(p, 40, 10), true).ShowPage("instance-list-view")
}

func (u *Ui) removeModal() {
	u.rootPage.RemovePage("modal")
}

func NewUi() *Ui {
	ec2svc := aws.NewEc2Svc()

	return &Ui{
		app:      tview.NewApplication(),
		rootPage: tview.NewPages(),
		ec2svc:   ec2svc,
	}
}

func modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}
