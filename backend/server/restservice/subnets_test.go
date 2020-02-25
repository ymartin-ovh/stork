package restservice

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	dbmodel "isc.org/stork/server/database/model"
	dbtest "isc.org/stork/server/database/test"
	dhcp "isc.org/stork/server/gen/restapi/operations/d_h_c_p"
	storktest "isc.org/stork/server/test"
)

// Check getting subnets via rest api functions.
func TestGetSubnets(t *testing.T) {
	db, dbSettings, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	settings := RestAPISettings{}
	fa := storktest.NewFakeAgents(nil)
	rapi, err := NewRestAPI(&settings, dbSettings, db, fa)
	require.NoError(t, err)
	ctx := context.Background()

	// get empty list of subnets
	params := dhcp.GetSubnetsParams{}
	rsp := rapi.GetSubnets(ctx, params)
	require.IsType(t, &dhcp.GetSubnetsOK{}, rsp)
	okRsp := rsp.(*dhcp.GetSubnetsOK)
	require.Len(t, okRsp.Payload.Items, 0)
	require.Equal(t, int64(0), okRsp.Payload.Total)

	// add machine
	m := &dbmodel.Machine{
		Address:   "localhost",
		AgentPort: 8080,
	}
	err = dbmodel.AddMachine(db, m)
	require.NoError(t, err)

	// add app kea with dhcp4 to machine
	var accessPoints []dbmodel.AccessPoint
	accessPoints = dbmodel.AppendAccessPoint(accessPoints, "control", "", "", 1114)

	a4 := &dbmodel.App{
		ID:           0,
		MachineID:    m.ID,
		Type:         dbmodel.KeaAppType,
		Active:       true,
		AccessPoints: accessPoints,
		Details: dbmodel.AppKea{
			Daemons: []*dbmodel.KeaDaemon{{
				Config: dbmodel.NewKeaConfig(&map[string]interface{}{
					"Dhcp4": &map[string]interface{}{
						"subnet4": []map[string]interface{}{{
							"id":     1,
							"subnet": "192.168.0.0/24",
							"pools": []map[string]interface{}{{
								"pool": "192.168.0.1-192.168.0.100",
							}, {
								"pool": "192.168.0.150-192.168.0.200",
							}},
						}},
					},
				}),
			}},
		},
	}
	err = dbmodel.AddApp(db, a4)
	require.NoError(t, err)

	appSubnets := []dbmodel.Subnet{
		{
			Prefix: "192.168.0.0/24",
			AddressPools: []dbmodel.AddressPool{
				{
					LowerBound: "192.168.0.1",
					UpperBound: "192.168.0.100",
				},
				{
					LowerBound: "192.168.0.150",
					UpperBound: "192.168.0.200",
				},
			},
		},
	}

	err = dbmodel.CommitNetworksIntoDB(db, []dbmodel.SharedNetwork{}, appSubnets, a4)
	require.NoError(t, err)

	// add app kea with dhcp6 to machine
	accessPoints = []dbmodel.AccessPoint{}
	accessPoints = dbmodel.AppendAccessPoint(accessPoints, "control", "", "", 1116)

	a6 := &dbmodel.App{
		ID:           0,
		MachineID:    m.ID,
		Type:         dbmodel.KeaAppType,
		Active:       true,
		AccessPoints: accessPoints,
		Details: dbmodel.AppKea{
			Daemons: []*dbmodel.KeaDaemon{{
				Config: dbmodel.NewKeaConfig(&map[string]interface{}{
					"Dhcp6": &map[string]interface{}{
						"subnet6": []map[string]interface{}{{
							"id":     2,
							"subnet": "2001:db8:1::/64",
							"pools":  []map[string]interface{}{},
						}},
					},
				}),
			}},
		},
	}
	err = dbmodel.AddApp(db, a6)
	require.NoError(t, err)

	appSubnets = []dbmodel.Subnet{
		{
			Prefix: "2001:db8:1::/64",
		},
	}
	err = dbmodel.CommitNetworksIntoDB(db, []dbmodel.SharedNetwork{}, appSubnets, a6)
	require.NoError(t, err)

	// add app kea with dhcp4 and dhcp6 to machine
	accessPoints = []dbmodel.AccessPoint{}
	accessPoints = dbmodel.AppendAccessPoint(accessPoints, "control", "", "", 1146)

	a46 := &dbmodel.App{
		ID:           0,
		MachineID:    m.ID,
		Type:         dbmodel.KeaAppType,
		Active:       true,
		AccessPoints: accessPoints,
		Details: dbmodel.AppKea{
			Daemons: []*dbmodel.KeaDaemon{{
				Config: dbmodel.NewKeaConfig(&map[string]interface{}{
					"Dhcp4": &map[string]interface{}{
						"subnet4": []map[string]interface{}{{
							"id":     3,
							"subnet": "192.118.0.0/24",
							"pools": []map[string]interface{}{{
								"pool": "192.118.0.1-192.118.0.200",
							}},
						}},
					},
				}),
			}, {
				Config: dbmodel.NewKeaConfig(&map[string]interface{}{
					"Dhcp6": &map[string]interface{}{
						"subnet6": []map[string]interface{}{{
							"id":     4,
							"subnet": "3001:db8:1::/64",
							"pools": []map[string]interface{}{{
								"pool": "3001:db8:1::/80",
							}},
						}},
						"shared-networks": []map[string]interface{}{{
							"name": "fox",
							"subnet6": []map[string]interface{}{{
								"id":     21,
								"subnet": "5001:db8:1::/64",
							}},
						}},
					},
				}),
			}},
		},
	}
	err = dbmodel.AddApp(db, a46)
	require.NoError(t, err)

	appNetworks := []dbmodel.SharedNetwork{
		{
			Name:   "fox",
			Family: 6,
			Subnets: []dbmodel.Subnet{
				{
					Prefix: "5001:db8:1::/64",
				},
			},
		},
	}

	appSubnets = []dbmodel.Subnet{
		{
			Prefix: "192.118.0.0/24",
			AddressPools: []dbmodel.AddressPool{
				{
					LowerBound: "192.118.0.1",
					UpperBound: "192.118.0.200",
				},
			},
		},
		{
			Prefix: "3001:db8:1::/64",
			AddressPools: []dbmodel.AddressPool{
				{
					LowerBound: "3001:db8:1::",
					UpperBound: "3001:db8:1:0:ffff::ffff",
				},
			},
		},
	}
	err = dbmodel.CommitNetworksIntoDB(db, appNetworks, appSubnets, a46)
	require.NoError(t, err)

	// get all subnets
	params = dhcp.GetSubnetsParams{}
	rsp = rapi.GetSubnets(ctx, params)
	require.IsType(t, &dhcp.GetSubnetsOK{}, rsp)
	okRsp = rsp.(*dhcp.GetSubnetsOK)
	require.Len(t, okRsp.Payload.Items, 5)
	require.Equal(t, int64(5), okRsp.Payload.Total)
	for _, sn := range okRsp.Payload.Items {
		switch sn.ID {
		case 1:
			require.Len(t, sn.Pools, 2)
		case 2:
			require.Len(t, sn.Pools, 0)
		case 21:
			require.Len(t, sn.Pools, 0)
		default:
			require.Len(t, sn.Pools, 1)
		}
	}

	// get subnets from app a4
	params = dhcp.GetSubnetsParams{
		AppID: &a4.ID,
	}
	rsp = rapi.GetSubnets(ctx, params)
	require.IsType(t, &dhcp.GetSubnetsOK{}, rsp)
	okRsp = rsp.(*dhcp.GetSubnetsOK)
	require.Len(t, okRsp.Payload.Items, 1)
	require.Equal(t, int64(1), okRsp.Payload.Total)
	require.Equal(t, a4.ID, okRsp.Payload.Items[0].AppID)
	require.Equal(t, int64(1), okRsp.Payload.Items[0].ID)

	// get subnets from app a46
	params = dhcp.GetSubnetsParams{
		AppID: &a46.ID,
	}
	rsp = rapi.GetSubnets(ctx, params)
	require.IsType(t, &dhcp.GetSubnetsOK{}, rsp)
	okRsp = rsp.(*dhcp.GetSubnetsOK)
	require.Len(t, okRsp.Payload.Items, 3)
	require.Equal(t, int64(3), okRsp.Payload.Total)
	// checking if returned subnet-ids have expected values
	require.ElementsMatch(t, []int64{3, 4, 21}, []int64{okRsp.Payload.Items[0].ID, okRsp.Payload.Items[1].ID, okRsp.Payload.Items[2].ID})

	// get v4 subnets
	var dhcpVer int64 = 4
	params = dhcp.GetSubnetsParams{
		DhcpVersion: &dhcpVer,
	}
	rsp = rapi.GetSubnets(ctx, params)
	require.IsType(t, &dhcp.GetSubnetsOK{}, rsp)
	okRsp = rsp.(*dhcp.GetSubnetsOK)
	require.Len(t, okRsp.Payload.Items, 2)
	require.Equal(t, int64(2), okRsp.Payload.Total)
	// checking if returned subnet-ids have expected values
	require.True(t, (okRsp.Payload.Items[0].ID == 1 && okRsp.Payload.Items[1].ID == 3) ||
		(okRsp.Payload.Items[0].ID == 3 && okRsp.Payload.Items[1].ID == 1))

	// get v6 subnets
	dhcpVer = 6
	params = dhcp.GetSubnetsParams{
		DhcpVersion: &dhcpVer,
	}
	rsp = rapi.GetSubnets(ctx, params)
	require.IsType(t, &dhcp.GetSubnetsOK{}, rsp)
	okRsp = rsp.(*dhcp.GetSubnetsOK)
	require.Len(t, okRsp.Payload.Items, 3)
	require.Equal(t, int64(3), okRsp.Payload.Total)
	// checking if returned subnet-ids have expected values
	require.ElementsMatch(t, []int64{2, 4, 21}, []int64{okRsp.Payload.Items[0].ID, okRsp.Payload.Items[1].ID, okRsp.Payload.Items[2].ID})

	// get subnets by text '118.0.0/2'
	text := "118.0.0/2"
	params = dhcp.GetSubnetsParams{
		Text: &text,
	}
	rsp = rapi.GetSubnets(ctx, params)
	require.IsType(t, &dhcp.GetSubnetsOK{}, rsp)
	okRsp = rsp.(*dhcp.GetSubnetsOK)
	require.Len(t, okRsp.Payload.Items, 1)
	require.Equal(t, int64(1), okRsp.Payload.Total)
	require.Equal(t, a46.ID, okRsp.Payload.Items[0].AppID)
	// checking if returned subnet-ids have expected values
	require.Equal(t, int64(3), okRsp.Payload.Items[0].ID)

	// get subnets by text '0.150-192.168'
	text = "0.150-192.168"
	params = dhcp.GetSubnetsParams{
		Text: &text,
	}
	rsp = rapi.GetSubnets(ctx, params)
	require.IsType(t, &dhcp.GetSubnetsOK{}, rsp)
	okRsp = rsp.(*dhcp.GetSubnetsOK)
	require.Len(t, okRsp.Payload.Items, 1)
	require.Equal(t, int64(1), okRsp.Payload.Total)
	require.Equal(t, a4.ID, okRsp.Payload.Items[0].AppID)
	// checking if returned subnet-ids have expected values
	require.Equal(t, int64(1), okRsp.Payload.Items[0].ID)
}

// Check getting shared networks via rest api functions.
func TestGetSharedNetworks(t *testing.T) {
	db, dbSettings, teardown := dbtest.SetupDatabaseTestCase(t)
	defer teardown()

	settings := RestAPISettings{}
	fa := storktest.NewFakeAgents(nil)
	rapi, err := NewRestAPI(&settings, dbSettings, db, fa)
	require.NoError(t, err)
	ctx := context.Background()

	// get empty list of subnets
	params := dhcp.GetSharedNetworksParams{}
	rsp := rapi.GetSharedNetworks(ctx, params)
	require.IsType(t, &dhcp.GetSharedNetworksOK{}, rsp)
	okRsp := rsp.(*dhcp.GetSharedNetworksOK)
	require.Len(t, okRsp.Payload.Items, 0)
	require.Equal(t, int64(0), okRsp.Payload.Total)

	// add machine
	m := &dbmodel.Machine{
		Address:   "localhost",
		AgentPort: 8080,
	}
	err = dbmodel.AddMachine(db, m)
	require.NoError(t, err)

	// add app kea with dhcp4 to machine
	var accessPoints []dbmodel.AccessPoint
	accessPoints = dbmodel.AppendAccessPoint(accessPoints, "control", "", "", 1114)

	a4 := &dbmodel.App{
		ID:           0,
		MachineID:    m.ID,
		Type:         dbmodel.KeaAppType,
		Active:       true,
		AccessPoints: accessPoints,
		Details: dbmodel.AppKea{
			Daemons: []*dbmodel.KeaDaemon{{
				Config: dbmodel.NewKeaConfig(&map[string]interface{}{
					"Dhcp4": &map[string]interface{}{
						"shared-networks": []map[string]interface{}{{
							"name": "frog",
							"subnet4": []map[string]interface{}{{
								"id":     11,
								"subnet": "192.1.0.0/24",
							}},
						}, {
							"name": "mouse",
							"subnet4": []map[string]interface{}{{
								"id":     12,
								"subnet": "192.2.0.0/24",
							}, {
								"id":     13,
								"subnet": "192.3.0.0/24",
							}},
						}},
					},
				}),
			}},
		},
	}
	err = dbmodel.AddApp(db, a4)
	require.NoError(t, err)

	appNetworks := []dbmodel.SharedNetwork{
		{
			Name:   "frog",
			Family: 4,
			Subnets: []dbmodel.Subnet{
				{
					Prefix: "192.1.0.0/24",
				},
			},
		},
		{
			Name:   "mouse",
			Family: 4,
			Subnets: []dbmodel.Subnet{
				{
					Prefix: "192.2.0.0/24",
				},
				{
					Prefix: "192.3.0.0/24",
				},
			},
		},
	}

	err = dbmodel.CommitNetworksIntoDB(db, appNetworks, []dbmodel.Subnet{}, a4)
	require.NoError(t, err)

	// add app kea with dhcp6 to machine
	accessPoints = []dbmodel.AccessPoint{}
	accessPoints = dbmodel.AppendAccessPoint(accessPoints, "control", "", "", 1116)

	a6 := &dbmodel.App{
		ID:           0,
		MachineID:    m.ID,
		Type:         dbmodel.KeaAppType,
		Active:       true,
		AccessPoints: accessPoints,
		Details: dbmodel.AppKea{
			Daemons: []*dbmodel.KeaDaemon{{
				Config: dbmodel.NewKeaConfig(&map[string]interface{}{
					"Dhcp6": &map[string]interface{}{
						"shared-networks": []map[string]interface{}{{
							"name": "fox",
							"subnet6": []map[string]interface{}{{
								"id":     21,
								"subnet": "5001:db8:1::/64",
							}, {
								"id":     22,
								"subnet": "6001:db8:1::/64",
							}},
						}},
					},
				}),
			}},
		},
	}
	err = dbmodel.AddApp(db, a6)
	require.NoError(t, err)

	appNetworks = []dbmodel.SharedNetwork{
		{
			Name:   "fox",
			Family: 6,
			Subnets: []dbmodel.Subnet{
				{
					Prefix: "5001:db8:1::/64",
				},
				{
					Prefix: "6001:db8:1::/64",
				},
			},
		},
	}
	err = dbmodel.CommitNetworksIntoDB(db, appNetworks, []dbmodel.Subnet{}, a6)
	require.NoError(t, err)

	// get all shared networks
	params = dhcp.GetSharedNetworksParams{}
	rsp = rapi.GetSharedNetworks(ctx, params)
	require.IsType(t, &dhcp.GetSharedNetworksOK{}, rsp)
	okRsp = rsp.(*dhcp.GetSharedNetworksOK)
	require.Len(t, okRsp.Payload.Items, 3)
	require.Equal(t, int64(3), okRsp.Payload.Total)
	for _, net := range okRsp.Payload.Items {
		require.Contains(t, []string{"frog", "mouse", "fox"}, net.Name)
		switch net.Name {
		case "frog":
			require.Len(t, net.Subnets, 1)
		case "mouse":
			require.Len(t, net.Subnets, 2)
		case "fox":
			require.Len(t, net.Subnets, 2)
		}
	}

	// get shared networks from app a4
	params = dhcp.GetSharedNetworksParams{
		AppID: &a4.ID,
	}
	rsp = rapi.GetSharedNetworks(ctx, params)
	require.IsType(t, &dhcp.GetSharedNetworksOK{}, rsp)
	okRsp = rsp.(*dhcp.GetSharedNetworksOK)
	require.Len(t, okRsp.Payload.Items, 2)
	require.Equal(t, int64(2), okRsp.Payload.Total)
	require.Equal(t, a4.ID, okRsp.Payload.Items[0].AppID)
	require.Equal(t, a4.ID, okRsp.Payload.Items[1].AppID)
	require.ElementsMatch(t, []string{"mouse", "frog"}, []string{okRsp.Payload.Items[0].Name, okRsp.Payload.Items[1].Name})
}
