package server

import (
	"os"

	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	"isc.org/stork/server/agentcomm"
	"isc.org/stork/server/apps/kea"
	dbops "isc.org/stork/server/database"
	dbmodel "isc.org/stork/server/database/model"
	"isc.org/stork/server/restservice"
)

// Global Stork Server state
type StorkServer struct {
	DbSettings dbops.DatabaseSettings
	Db         *dbops.PgDB

	AgentsSettings agentcomm.AgentsSettings
	Agents         agentcomm.ConnectedAgents

	RestAPISettings restservice.RestAPISettings
	RestAPI         *restservice.RestAPI

	StatsPuller *kea.StatsPuller
	HostsPuller *kea.HostsPuller
}

func (ss *StorkServer) ParseArgs() {
	// Process command line flags.
	parser := flags.NewParser(nil, flags.Default) // TODO: change nil to some main group of server settings
	parser.ShortDescription = "Stork Server"
	parser.LongDescription = "Stork Server is a Kea and BIND 9 Dashboard"

	// Process Database specific args.
	_, err := parser.AddGroup("Database ConnectionFlags", "", &ss.DbSettings)
	if err != nil {
		log.Fatalf("FATAL error: %+v", err)
	}

	// Process ReST API specific args.
	_, err = parser.AddGroup("ReST Server Flags", "", &ss.RestAPISettings)
	if err != nil {
		log.Fatalf("FATAL error: %+v", err)
	}

	// Process agent comm specific args.
	_, err = parser.AddGroup("Agents Communication Flags", "", &ss.AgentsSettings)
	if err != nil {
		log.Fatalf("FATAL error: %+v", err)
	}

	// Do args parsing.
	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}
}

// Init for Stork Server state
func NewStorkServer() (ss *StorkServer, err error) {
	ss = &StorkServer{}
	ss.ParseArgs()

	// setup connected agents
	ss.Agents = agentcomm.NewConnectedAgents(&ss.AgentsSettings)
	defer func() {
		if err != nil {
			ss.Agents.Shutdown()
		}
	}()

	// setup database connection
	ss.Db, err = dbops.NewPgDB(&ss.DbSettings)
	if err != nil {
		return nil, err
	}

	// initialize stork settings
	err = dbmodel.InitializeSettings(ss.Db)
	if err != nil {
		return nil, err
	}

	// setup kea stats puller
	ss.StatsPuller, err = kea.NewStatsPuller(ss.Db, ss.Agents)
	if err != nil {
		return nil, err
	}

	// Setup Kea hosts puller.
	ss.HostsPuller, err = kea.NewHostsPuller(ss.Db, ss.Agents)
	if err != nil {
		return nil, err
	}

	// setup ReST API service
	r, err := restservice.NewRestAPI(&ss.RestAPISettings, &ss.DbSettings, ss.Db, ss.Agents)
	if err != nil {
		ss.StatsPuller.Shutdown()
		ss.Db.Close()
		return nil, err
	}
	ss.RestAPI = r
	return ss, nil
}

// Run Stork Server
func (ss *StorkServer) Serve() {
	// Start listening for requests from ReST API.
	err := ss.RestAPI.Serve()
	if err != nil {
		log.Fatalf("FATAL error: %+v", err)
	}
}

// Shutdown for Stork Server state
func (ss *StorkServer) Shutdown() {
	log.Println("Shutting down Stork Server")
	ss.RestAPI.Shutdown()
	ss.StatsPuller.Shutdown()
	ss.Db.Close()
	ss.Agents.Shutdown()
	log.Println("Stork Server shut down")
}
