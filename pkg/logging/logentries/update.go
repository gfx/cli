package logentries

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// UpdateCommand calls the Fastly API to update Logentries logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int

	// optional
	NewName           common.OptionalString
	Port              common.OptionalUint
	UseTLS            common.OptionalBool
	Token             common.OptionalString
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	ResponseCondition common.OptionalString
	Placement         common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update a Logentries logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the Logentries logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the Logentries logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("port", "The port number").Action(c.Port.Set).UintVar(&c.Port.Value)
	c.CmdClause.Flag("use-tls", "Whether to use TLS for secure logging. Can be either true or false").Action(c.UseTLS.Set).BoolVar(&c.UseTLS.Value)
	c.CmdClause.Flag("auth-token", "Use token based authentication (https://logentries.com/doc/input-token/)").Action(c.Token.Set).StringVar(&c.Token.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug. This field is not required and has no default value").Action(c.Placement.Set).StringVar(&c.Placement.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateLogentriesInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	logentries, err := c.Globals.Client.GetLogentries(&fastly.GetLogentriesInput{
		ServiceID:      serviceID,
		Name:           c.EndpointName,
		ServiceVersion: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateLogentriesInput{
		ServiceID:         logentries.ServiceID,
		ServiceVersion:    logentries.ServiceVersion,
		Name:              logentries.Name,
		NewName:           fastly.String(logentries.Name),
		Port:              fastly.Uint(logentries.Port),
		UseTLS:            fastly.CBool(logentries.UseTLS),
		Token:             fastly.String(logentries.Token),
		Format:            fastly.String(logentries.Format),
		FormatVersion:     fastly.Uint(logentries.FormatVersion),
		ResponseCondition: fastly.String(logentries.ResponseCondition),
		Placement:         fastly.String(logentries.Placement),
	}

	// Set new values if set by user.
	if c.NewName.Valid {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.Port.Valid {
		input.Port = fastly.Uint(c.Port.Value)
	}

	if c.UseTLS.Valid {
		input.UseTLS = fastly.CBool(c.UseTLS.Value)
	}

	if c.Token.Valid {
		input.Token = fastly.String(c.Token.Value)
	}

	if c.Format.Valid {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.ResponseCondition.Valid {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.Placement.Valid {
		input.Placement = fastly.String(c.Placement.Value)
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	logentries, err := c.Globals.Client.UpdateLogentries(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Logentries logging endpoint %s (service %s version %d)", logentries.Name, logentries.ServiceID, logentries.ServiceVersion)
	return nil
}
