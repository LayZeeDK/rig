package group

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	outputJSON bool
)

var (
	offset int
	limit  int
)

var (
	name string
)

type Cmd struct {
	fx.In

	Ctx context.Context
	Rig rig.Client
}

func (c Cmd) Setup(parent *cobra.Command) {
	group := &cobra.Command{
		Use:   "group",
		Short: "Manage user groups",
	}

	create := &cobra.Command{
		Use:               "create",
		Short:             "Create a new group",
		RunE:              c.create,
		Args:              cobra.NoArgs,
		ValidArgsFunction: common.NoCompletions,
	}
	create.Flags().StringVarP(&name, "name", "n", "", "name of the group")
	create.RegisterFlagCompletionFunc("name", common.NoCompletions)
	group.AddCommand(create)

	delete := &cobra.Command{
		Use:               "delete [group-id | group-name]",
		Short:             "Delete a group",
		RunE:              c.delete,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(c.completions, common.MaxArgsCompletionFilter(1)),
	}
	group.AddCommand(delete)

	update := &cobra.Command{
		Use:               "update [group-id | group-name]",
		Short:             "Update a group",
		Args:              cobra.MaximumNArgs(1),
		RunE:              c.update,
		ValidArgsFunction: common.Complete(c.completions, common.MaxArgsCompletionFilter(1)),
	}
	group.AddCommand(update)

	get := &cobra.Command{
		Use:               "get [group-id | group-name]",
		Short:             "Get one or multiple groups",
		Args:              cobra.MaximumNArgs(1),
		Aliases:           []string{"ls"},
		RunE:              c.get,
		ValidArgsFunction: common.Complete(c.completions, common.MaxArgsCompletionFilter(1)),
	}
	get.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	get.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	get.Flags().IntVarP(&offset, "offset", "o", 0, "offset the number of groups to return")
	get.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	get.RegisterFlagCompletionFunc("limit", common.NoCompletions)
	get.RegisterFlagCompletionFunc("offset", common.NoCompletions)
	group.AddCommand(get)

	getMembers := &cobra.Command{
		Use:               "get-members [group-id | group-name]",
		Short:             "Get members of a group",
		RunE:              c.listMembers,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(c.completions, common.MaxArgsCompletionFilter(1)),
	}
	getMembers.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	getMembers.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of members to return")
	getMembers.Flags().IntVarP(&offset, "offset", "o", 0, "offset the number of members to return")
	getMembers.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	getMembers.RegisterFlagCompletionFunc("limit", common.NoCompletions)
	getMembers.RegisterFlagCompletionFunc("offset", common.NoCompletions)
	group.AddCommand(getMembers)

	getGroupsForUser := &cobra.Command{
		Use:               "get-groups-for-user [user-id | {email|username|phone}]",
		Short:             "Get groups that a user is a member of",
		RunE:              c.listGroupsForUser,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(c.userCompletions, common.MaxArgsCompletionFilter(1)),
	}
	getGroupsForUser.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	getGroupsForUser.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	getGroupsForUser.Flags().IntVarP(&offset, "offset", "o", 0, "offset the number of groups to return")
	getGroupsForUser.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	getGroupsForUser.RegisterFlagCompletionFunc("limit", common.NoCompletions)
	getGroupsForUser.RegisterFlagCompletionFunc("offset", common.NoCompletions)
	group.AddCommand(getGroupsForUser)

	parent.AddCommand(group)
}

func (c Cmd) completions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	completions := []string{}
	resp, err := c.Rig.Group().List(c.Ctx, &connect.Request[group.ListRequest]{
		Msg: &group.ListRequest{},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, g := range resp.Msg.GetGroups() {
		if strings.HasPrefix(g.GetName(), toComplete) {
			completions = append(completions, formatGroup(g))
		}
	}

	if len(completions) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func formatGroup(g *group.Group) string {
	return fmt.Sprintf("%s\t (#Members: %v)", g.GetName(), g.GetNumMembers())
}

func (c Cmd) userCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	completions := []string{}
	resp, err := c.Rig.User().List(c.Ctx, &connect.Request[user.ListRequest]{
		Msg: &user.ListRequest{},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, u := range resp.Msg.GetUsers() {
		if strings.HasPrefix(u.GetPrintableName(), toComplete) {
			completions = append(completions, formatUser(u))
		}
	}

	if len(completions) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func formatUser(u *model.UserEntry) string {
	return fmt.Sprintf("%s\t (ID: %s)", u.GetPrintableName(), u.GetUserId())
}
