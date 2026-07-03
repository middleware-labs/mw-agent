package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.uber.org/multierr"
)

type MongoHost struct {
	Endpoint string `mapstructure:"endpoint"`
}

type MongodbReceiver struct {
	Hosts            []MongoHost `mapstructure:"hosts"`
	Scheme           string      `mapstructure:"scheme"`
	Username         string      `mapstructure:"username"`
	Password         string      `mapstructure:"password"`
	ReplicaSet       string      `mapstructure:"replica_set,omitempty"`
	DirectConnection bool        `mapstructure:"direct_connection"`
}

func (r *MongodbReceiver) Validate() error {
	if len(r.Hosts) == 0 {
		return errors.New("mongodb: no hosts configured")
	}

	var err error
	for _, h := range r.Hosts {
		if h.Endpoint == "" {
			err = multierr.Append(err, errors.New("mongodb: empty endpoint in hosts"))
		}
	}

	if r.Scheme != "" && r.Scheme != "mongodb" && r.Scheme != "mongodb+srv" {
		err = multierr.Append(err, fmt.Errorf("mongodb: invalid scheme %q", r.Scheme))
	}

	if r.Scheme == "mongodb+srv" && len(r.Hosts) != 1 {
		err = multierr.Append(err, errors.New("mongodb: mongodb+srv requires exactly one host"))
	}

	if r.Username != "" && r.Password == "" {
		err = multierr.Append(err, errors.New("mongodb: username provided without password"))
	} else if r.Username == "" && r.Password != "" {
		err = multierr.Append(err, errors.New("mongodb: password provided without username"))
	}

	return err
}

func (r *MongodbReceiver) connect(ctx context.Context) (*mongo.Client, error) {
	scheme := r.Scheme
	if scheme == "" {
		scheme = "mongodb"
	}

	var hostlist []string
	for _, h := range r.Hosts {
		hostlist = append(hostlist, h.Endpoint)
	}

	uri := scheme + "://" + strings.Join(hostlist, ",")

	opts := options.Client().ApplyURI(uri)
	opts.SetConnectTimeout(5 * time.Second)

	if r.Username != "" && r.Password != "" {
		opts.SetAuth(options.Credential{
			Username:   r.Username,
			Password:   r.Password,
			AuthSource: "admin",
		})
	}

	if r.ReplicaSet != "" {
		opts.SetReplicaSet(r.ReplicaSet)
	}

	if r.DirectConnection {
		opts.SetDirect(true)
	}

	return mongo.Connect(opts)
}

func (r *MongodbReceiver) CheckHealth(ctx context.Context) error {
	checks, err := r.CheckHealthDetailed(ctx)
	if err != nil {
		return err
	}

	var missing []string
	for _, c := range checks {
		if !c.Granted {
			missing = append(missing, c.Name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("mongodb: missing permissions: %s", strings.Join(missing, ", "))
	}

	return nil
}

func (r *MongodbReceiver) CheckHealthDetailed(ctx context.Context) ([]PermissionCheck, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, err := r.connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("mongodb: failed to connect: %w", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("mongodb: ping failed: %w", err)
	}

	return r.checkPermissions(ctx, client), nil
}

func (r *MongodbReceiver) checkPermissions(ctx context.Context, client *mongo.Client) []PermissionCheck {
	checks := []PermissionCheck{
		{Name: "CONNECTED", Granted: true},
		{Name: "AUTHENTICATED", Granted: true},
		{Name: "clusterMonitor"},
		{Name: "dbAdminAnyDatabase"},
		{Name: "profiling"},
	}

	var connStatus bson.M
	err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "connectionStatus", Value: 1}}).Decode(&connStatus)
	if err != nil {
		for i := 2; i < len(checks); i++ {
			checks[i].Err = err
		}
		return checks
	}

	roles := extractRoles(connStatus)
	checks[2].Granted = roles["clusterMonitor"]
	checks[3].Granted = roles["dbAdminAnyDatabase"]
	if !checks[2].Granted {
		checks[2].Err = fmt.Errorf("role not granted to user")
	}
	if !checks[3].Granted {
		checks[3].Err = fmt.Errorf("role not granted to user")
	}

	var profileStatus bson.M
	err = client.Database("admin").RunCommand(ctx, bson.D{{Key: "profile", Value: -1}}).Decode(&profileStatus)
	if err != nil {
		checks[4].Err = err
		return checks
	}

	if was, ok := profileStatus["was"]; ok {
		level, _ := was.(int32)
		checks[4].Granted = level > 0
		if level == 0 {
			checks[4].Err = fmt.Errorf("profiling is disabled (level 0)")
		}
	}

	return checks
}

func extractRoles(connStatus bson.M) map[string]bool {
	result := make(map[string]bool)

	authInfo, _ := connStatus["authInfo"]
	if authInfo == nil {
		return result
	}

	var roles bson.A
	switch ai := authInfo.(type) {
	case bson.D:
		for _, elem := range ai {
			if elem.Key == "authenticatedUserRoles" {
				roles, _ = elem.Value.(bson.A)
			}
		}
	case bson.M:
		roles, _ = ai["authenticatedUserRoles"].(bson.A)
	}

	for _, r := range roles {
		switch role := r.(type) {
		case bson.D:
			for _, field := range role {
				if field.Key == "role" {
					if name, ok := field.Value.(string); ok {
						result[name] = true
					}
				}
			}
		case bson.M:
			if name, ok := role["role"].(string); ok {
				result[name] = true
			}
		}
	}

	return result
}
