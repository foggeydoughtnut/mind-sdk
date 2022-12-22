package SmartPathPlanner

import (
	"mind/core/framework/skill"
)

type SmartPathPlanner struct {
	skill.Base
}

func NewSkill() skill.Interface {
	// Use this method to create a new skill.

	return &SmartPathPlanner{}
}

func (d *SmartPathPlanner) OnStart() {
	// Use this method to do something when this skill is starting.
}

func (d *SmartPathPlanner) OnClose() {
	// Use this method to do something when this skill is closing.
}

func (d *SmartPathPlanner) OnConnect() {
	// Use this method to do something when the remote connected.
}

func (d *SmartPathPlanner) OnDisconnect() {
	// Use this method to do something when the remote disconnected.
}

func (d *SmartPathPlanner) OnRecvJSON(data []byte) {
	// Use this method to do something when skill receive json data from remote client.
}

func (d *SmartPathPlanner) OnRecvString(data string) {
	// Use this method to do something when skill receive string from remote client.
}
