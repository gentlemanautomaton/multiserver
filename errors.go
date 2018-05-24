package multiserver

import "errors"

// ErrGroupStarted is returned by the Group's ListenAndServe method if
// the group has already been started.
var ErrGroupStarted = errors.New("multiserver: server group already started")
