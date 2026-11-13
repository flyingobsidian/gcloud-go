package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	networkconnectivity "google.golang.org/api/networkconnectivity/v1"
)

// --- hubs (#901) ---

func (grp *networkConnectivityCmds) registerNCHubs(root *cobra.Command) {
	crud := &ncCRUD[networkconnectivity.Hub]{
		s:     grp,
		group: "hubs", singular: "hub", collection: "global/hubs",
		parentFn: func() (string, error) { return grp.ncGlobalParent() },
		createFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, id string, body *networkconnectivity.Hub, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.Global.Hubs.Create(parent, body).HubId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networkconnectivity.Service, name, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.Global.Hubs.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networkconnectivity.Service, name string) (*networkconnectivity.Hub, error) {
			return svc.Projects.Locations.Global.Hubs.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, pageToken, filter string) ([]*networkconnectivity.Hub, string, error) {
			c := svc.Projects.Locations.Global.Hubs.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			if filter != "" {
				c = c.Filter(filter)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.Hubs, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networkconnectivity.Service, name string, body *networkconnectivity.Hub, mask, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.Global.Hubs.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(h *networkconnectivity.Hub) string { return ncBasename(h.Name) },
		secondaryCol:    func(h *networkconnectivity.Hub) string { return h.State },
		secondaryHeader: "STATE",
	}

	// Hubs live under projects/PROJECT/locations/global by construction, so no
	// --location flag is exposed; the parentFn hardcodes it.
	group := crud.build(root, "hubs", "Manage hubs", func(_ ...*cobra.Command) {}, true)

	// accept-spoke / reject-spoke / list-spokes
	accept := &cobra.Command{
		Use: "accept-spoke HUB", Short: "Accept a spoke into a hub",
		Args: cobra.ExactArgs(1), RunE: grp.runNCHubAcceptSpoke,
	}
	reject := &cobra.Command{
		Use: "reject-spoke HUB", Short: "Reject a spoke from a hub",
		Args: cobra.ExactArgs(1), RunE: grp.runNCHubRejectSpoke,
	}
	listSpokes := &cobra.Command{
		Use: "list-spokes HUB", Short: "List spokes in a hub",
		Args: cobra.ExactArgs(1), RunE: grp.runNCHubListSpokes,
	}
	accept.Flags().StringVar(&grp.flagNCSpokeURI, "spoke", "", "URI of the spoke (required)")
	_ = accept.MarkFlagRequired("spoke")
	reject.Flags().StringVar(&grp.flagNCSpokeURI, "spoke", "", "URI of the spoke (required)")
	_ = reject.MarkFlagRequired("spoke")
	reject.Flags().StringVar(&grp.flagNCDetails, "details", "", "Optional information from the hub administrator")
	grp.addNCAsyncFlag(accept, reject)
	grp.addNCRequestIDFlag(accept, reject)
	grp.addNCFilterFlag(listSpokes)
	grp.addNCFormatFlag(listSpokes)

	group.AddCommand(accept, reject, listSpokes)
}

func (grp *networkConnectivityCmds) ncResolveHub(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("hub is required")
	}
	parent, err := grp.ncGlobalParent()
	if err != nil {
		return "", err
	}
	return ncChild(parent, "hubs", id), nil
}

func (grp *networkConnectivityCmds) runNCHubAcceptSpoke(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	name, err := grp.ncResolveHub(args[0])
	if err != nil {
		return err
	}
	req := &networkconnectivity.AcceptHubSpokeRequest{SpokeUri: grp.flagNCSpokeURI, RequestId: grp.flagNCRequestID}
	op, err := svc.Projects.Locations.Global.Hubs.AcceptSpoke(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("accepting spoke: %w", err)
	}
	return grp.ncFinishOp(ctx, svc, op, "Accept spoke", args[0])
}

func (grp *networkConnectivityCmds) runNCHubRejectSpoke(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	name, err := grp.ncResolveHub(args[0])
	if err != nil {
		return err
	}
	req := &networkconnectivity.RejectHubSpokeRequest{SpokeUri: grp.flagNCSpokeURI, Details: grp.flagNCDetails, RequestId: grp.flagNCRequestID}
	op, err := svc.Projects.Locations.Global.Hubs.RejectSpoke(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rejecting spoke: %w", err)
	}
	return grp.ncFinishOp(ctx, svc, op, "Reject spoke", args[0])
}

func (grp *networkConnectivityCmds) runNCHubListSpokes(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	name, err := grp.ncResolveHub(args[0])
	if err != nil {
		return err
	}
	var all []*networkconnectivity.Spoke
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.Hubs.ListSpokes(name).Context(ctx)
		if grp.flagNCFilter != "" {
			call = call.Filter(grp.flagNCFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing hub spokes: %w", err)
		}
		all = append(all, resp.Spokes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagNCFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagNCFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-50s %s\n", "NAME", "STATE")
	for _, s := range all {
		fmt.Fprintf(grp.inv.stdout, "%-50s %s\n", ncBasename(s.Name), s.State)
	}
	return nil
}

// --- spokes (#910) ---

func (grp *networkConnectivityCmds) registerNCSpokes(root *cobra.Command) {
	crud := &ncCRUD[networkconnectivity.Spoke]{
		s:     grp,
		group: "spokes", singular: "spoke", collection: "spokes",
		parentFn: grp.ncProjectParent,
		createFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, id string, body *networkconnectivity.Spoke, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.Spokes.Create(parent, body).SpokeId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networkconnectivity.Service, name, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.Spokes.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networkconnectivity.Service, name string) (*networkconnectivity.Spoke, error) {
			return svc.Projects.Locations.Spokes.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, pageToken, filter string) ([]*networkconnectivity.Spoke, string, error) {
			c := svc.Projects.Locations.Spokes.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			if filter != "" {
				c = c.Filter(filter)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.Spokes, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networkconnectivity.Service, name string, body *networkconnectivity.Spoke, mask, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.Spokes.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(s *networkconnectivity.Spoke) string { return ncBasename(s.Name) },
		secondaryCol:    func(s *networkconnectivity.Spoke) string { return s.State },
		secondaryHeader: "STATE",
	}
	group := crud.build(root, "spokes", "Manage spokes", grp.addNCLocationFlag, true)

	// accept / reject: these hit the hub's AcceptSpoke/RejectSpoke endpoint
	// with the resolved spoke URI as the payload.
	accept := &cobra.Command{
		Use: "accept SPOKE", Short: "Accept a spoke proposal on its hub",
		Args: cobra.ExactArgs(1), RunE: grp.runNCSpokeAccept,
	}
	reject := &cobra.Command{
		Use: "reject SPOKE", Short: "Reject a spoke proposal on its hub",
		Args: cobra.ExactArgs(1), RunE: grp.runNCSpokeReject,
	}
	accept.Flags().StringVar(&grp.flagNCHub, "hub", "", "Hub URI or short name (required)")
	_ = accept.MarkFlagRequired("hub")
	reject.Flags().StringVar(&grp.flagNCHub, "hub", "", "Hub URI or short name (required)")
	_ = reject.MarkFlagRequired("hub")
	reject.Flags().StringVar(&grp.flagNCDetails, "details", "", "Optional information from the hub administrator")
	grp.addNCLocationFlag(accept, reject)
	grp.addNCAsyncFlag(accept, reject)
	grp.addNCRequestIDFlag(accept, reject)
	group.AddCommand(accept, reject)
}

func (grp *networkConnectivityCmds) ncResolveSpoke(id string) (string, error) {
	parent, err := grp.ncProjectParent()
	if err != nil {
		return "", err
	}
	return ncChild(parent, "spokes", id), nil
}

func (grp *networkConnectivityCmds) runNCSpokeAccept(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	spokeURI, err := grp.ncResolveSpoke(args[0])
	if err != nil {
		return err
	}
	hubName, err := grp.ncResolveHub(grp.flagNCHub)
	if err != nil {
		return err
	}
	req := &networkconnectivity.AcceptHubSpokeRequest{SpokeUri: spokeURI, RequestId: grp.flagNCRequestID}
	op, err := svc.Projects.Locations.Global.Hubs.AcceptSpoke(hubName, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("accepting spoke: %w", err)
	}
	return grp.ncFinishOp(ctx, svc, op, "Accept spoke", args[0])
}

func (grp *networkConnectivityCmds) runNCSpokeReject(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	spokeURI, err := grp.ncResolveSpoke(args[0])
	if err != nil {
		return err
	}
	hubName, err := grp.ncResolveHub(grp.flagNCHub)
	if err != nil {
		return err
	}
	req := &networkconnectivity.RejectHubSpokeRequest{SpokeUri: spokeURI, Details: grp.flagNCDetails, RequestId: grp.flagNCRequestID}
	op, err := svc.Projects.Locations.Global.Hubs.RejectSpoke(hubName, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rejecting spoke: %w", err)
	}
	return grp.ncFinishOp(ctx, svc, op, "Reject spoke", args[0])
}
