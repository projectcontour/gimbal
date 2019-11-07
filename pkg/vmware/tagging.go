package vmware

import (
	"context"
	"log"
	"net/url"
	"os"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/soap"
)

// TaggingClient is the tagging client used to tag a VM
type TaggingClient struct {
	vc *govmomi.Client
	// rest *rest.TaggingClient
	manager *tags.Manager
	log     Logger
}

// Logger is an interface to make logging configurable
type Logger interface {
	Printf(format string, v ...interface{})
}

// New creates a tagging client with a custom logger. If logger is nil, log.Logger will be used.
func NewTaggingClient(ctx context.Context, logger Logger, vc string, vcuser string, vcpass string, insecure bool) (*TaggingClient, error) {
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}
	u, err := soap.ParseURL(vc)
	if err != nil {
		log.Printf("could not parse vCenter client URL: %v", err)
		return nil, err
	}

	u.User = url.UserPassword(vcuser, vcpass)
	c, err := govmomi.NewClient(ctx, u, insecure)
	if err != nil {
		log.Printf("could not get vCenter client: %v", err)
		return nil, err
	}

	r := rest.NewClient(c.Client)
	err = r.Login(ctx, u.User)
	if err != nil {
		log.Printf("could not get VAPI REST client: %v", err)
		return nil, err
	}
	tm := tags.NewManager(r)

	return &TaggingClient{vc: c, manager: tm, log: logger}, nil
}
