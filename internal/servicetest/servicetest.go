package servicetest

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	proto "github.com/golang/protobuf/proto"
	ptypes "github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/jamiealquiza/envy"
	nats "github.com/nats-io/go-nats"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/Xawery/auth-service/account"
	"github.com/Xawery/auth-service/authn"
	"github.com/Xawery/auth-service/event"
	"github.com/Xawery/auth-service/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"gopkg.in/square/go-jose.v2/jwt"
)

var (
	testEndpoint    = flag.String("endpoint", "localhost:8080", "Backend Endpoint")
	testNatsURL     = flag.String("nats", "nats://localhost:4222", "NATS URL")
	testNatsCluster = flag.String("natsCluster", "test-cluster", "NATS Cluster name")
	testSecret      = flag.String("secret", "secret", "Auth secret")
)

var (
	Nats     *nats.Conn
	Health   health.HealthClient
	Accounts account.AccountsClient
	Authn    authn.AuthnClient
)

func TestMain(m *testing.M) {
	envy.Parse("TEST")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, *testEndpoint, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Failed to establish connection to service: %+v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	Nats, err = dialNats(ctx)
	if err != nil {
		fmt.Printf("Failed to establish connection to NATS: %+v\n", err)
		os.Exit(1)
	}
	defer Nats.Close()

	Health = health.NewHealthClient(conn)
	Accounts = account.NewAccountsClient(conn)
	Authn = authn.NewAuthnClient(conn)

	ret := m.Run()
	conn.Close()

	os.Exit(ret)
}

func dialNats(ctx context.Context) (*nats.Conn, error) {
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-tick.C:
			nc, err := nats.Connect(*testNatsURL)
			if err != nil {
				continue
			}

			return nc, nil
		}
	}
}

type EventMatcher func(map[string]interface{}) bool
type EventHandler func(tpe proto.Message, matchers ...EventMatcher) error

func EventSink(ctx context.Context, t *testing.T, queue string) EventHandler {
	conn, err := stan.Connect(*testNatsCluster, uuid.New().String(), stan.NatsConn(Nats))
	require.NoError(t, err, "Failed to establish connection to nats")

	ch := make(chan *event.Event)

	_, err = conn.Subscribe(queue, func(msg *stan.Msg) {
		e := event.Event{}
		if !assert.NoError(t, proto.Unmarshal(msg.Data, &e), "Failed to unmarshal event") {
			ch <- nil
		}

		ch <- &e
	})
	require.NoErrorf(t, err, "Failed to subscribe to %s queue", queue)

	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		}
	}()

	return func(tpe proto.Message, matchers ...EventMatcher) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case e := <-ch:
				if err := ptypes.UnmarshalAny(e.Payload, tpe); err != nil {
					continue
				}

				if eventMatches(tpe, matchers...) {
					return nil
				}
			}
		}
	}
}

func eventMatches(tpe proto.Message, matchers ...EventMatcher) bool {
	m := jsonpb.Marshaler{}
	buf := bytes.Buffer{}

	if err := m.Marshal(&buf, tpe); err != nil {
		return false
	}

	data := make(map[string]interface{})
	if err := json.NewDecoder(&buf).Decode(&data); err != nil {
		return false
	}

	for _, matcher := range matchers {
		if !matcher(data) {
			return false
		}
	}

	return true
}

func IDEquals(id int64) func(map[string]interface{}) bool {
	return func(data map[string]interface{}) bool {
		return data["id"] == strconv.FormatInt(id, 10)
	}
}

func UsernameEquals(username string) func(map[string]interface{}) bool {
	return func(data map[string]interface{}) bool {
		return data["username"] == username
	}
}

func Ready(ctx context.Context, t *testing.T) (*health.CheckReply, bool) {
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, assert.NoError(t, ctx.Err())
		case <-tick.C:
			reply, err := Health.Check(ctx, &empty.Empty{})
			if err != nil {
				continue
			}

			if reply.Status == health.CheckReply_READY {
				return reply, true
			}
		}
	}
}

func ParseToken(raw string, claims interface{}) error {
	tok, err := jwt.ParseSigned(raw)
	if err != nil {
		return err
	}

	key := []byte(*testSecret)

	jc := jwt.Claims{}
	if err := tok.Claims(key, &jc, claims); err != nil {
		return err
	}

	return jc.ValidateWithLeeway(jwt.Expected{}, 5*time.Second)
}
