package app

import (
	"runtime"
	"time"

	"github.com/robfig/revel"
	"github.com/zimmski/tirion/backend"
)

var Db backend.Backend

func init() {
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.ActionInvoker,           // Invoke the action.
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	revel.TemplateFuncs["ne"] = func(a, b interface{}) bool { return a != b }

	time.Local = time.UTC

	revel.OnAppStart(func() {
		var err error

		var dbDriver, _ = revel.Config.String("db.driver")
		var dbSpec, _ = revel.Config.String("db.spec")

		Db, err = backend.NewBackend(dbDriver)

		if err != nil {
			panic(err)
		}

		Db.Init(dbSpec)
	})
}
