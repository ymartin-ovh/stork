package storktest

import (
	"net/http"

	dbmodel "isc.org/stork/server/database/model"
)

// Helper struct to mock EventCenter behavior.
type FakeEventCenter struct {
}

func (fec *FakeEventCenter) AddInfoEvent(text string, objects ...interface{}) {
}
func (fec *FakeEventCenter) AddWarnEvent(text string, objects ...interface{}) {
}
func (fec *FakeEventCenter) AddErroEvent(text string, objects ...interface{}) {
}
func (fec *FakeEventCenter) AddEvent(event *dbmodel.Event) {
}
func (fec *FakeEventCenter) Shutdown() {
}
func (fec *FakeEventCenter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
}