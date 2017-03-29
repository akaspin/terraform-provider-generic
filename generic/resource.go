package generic

import (
	"github.com/hashicorp/terraform/helper/schema"
	"fmt"
	"math/rand"
	"time"
	"strings"
	"runtime"
	"context"
	"os/exec"
	"log"
	"os"
)

const (
	// maxBufSize limits how much output we collect from a local
	// invocation. This is to prevent TF memory usage from growing
	// to an enormous amount due to a faulty process.
	maxBufSize = 8 * 1024
)

func Resource() *schema.Resource  {
	return &schema.Resource{
		Create: resourceCreate,
		Update: resourceUpdate,
		Read:   resourceRead,
		Delete: resourceDelete,
		Schema: map[string]*schema.Schema {
			"trigger": {
				Description: "Triggers to force new resource",
				Type: schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"inline": {
				Type:          schema.TypeList,
				Elem:          &schema.Schema{Type: schema.TypeString},
				PromoteSingle: true,
				Required: true,

			},
			"phase": {
				Type: schema.TypeString,
				Default: "create",
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (r []string, errs []error) {
					switch v.(string) {
					case "create", "update", "destroy":
					default:
						errs = append(errs, fmt.Errorf("Invalid phase %s", v.(string)))
					}
					return
				},
			},
			"timeout": {
				Type: schema.TypeString,
				Optional: true,
				Default: "",
			},
			"ignore_errors": {
				Type: schema.TypeBool,
				Optional: true,
				Default: false,
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, meta interface{}) (err error) {
	d.SetId(fmt.Sprintf("%d", rand.Int()))
	if "create" == d.Get("phase").(string) {
		var runner *Runner
		runner, err = NewRunner(d.Get("timeout").(string), d.Get("ignore_errors").(bool), d.Get("inline").([]interface{}))
		if err != nil {
			return
		}
		err = runner.Run()
	}
	return
}

func resourceUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	if "update" == d.Get("phase").(string) {
		var runner *Runner
		runner, err = NewRunner(d.Get("timeout").(string), d.Get("ignore_errors").(bool), d.Get("inline").([]interface{}))
		if err != nil {
			return
		}
		err = runner.Run()
	}
	return
}

func resourceRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDelete(d *schema.ResourceData, meta interface{}) (err error) {
	d.SetId("")
	if "destroy" == d.Get("phase").(string) {
		var runner *Runner
		runner, err = NewRunner(d.Get("timeout").(string), d.Get("ignore_errors").(bool), d.Get("inline").([]interface{}))
		if err != nil {
			return
		}
		err = runner.Run()
	}
	return nil
}


type Runner struct {
	Timeout time.Duration
	IgnoreErrors bool
	Inline string
}

func NewRunner(timeout string, ignoreErrors bool, inline []interface{}) (r *Runner, err error) {
	r = &Runner{
		IgnoreErrors: ignoreErrors,
	}
	if timeout != "" {
		if r.Timeout, err = time.ParseDuration(timeout); err != nil {
			return
		}
	}
	var lines []string
	for _, line := range inline {
		lines = append(lines, line.(string))
	}
	lines = append(lines, "")
	r.Inline = strings.Join(lines, "\n")
	return
}

func (r *Runner) Run() (err error) {
	// Execute the command using a shell
	var shell, flag string
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/C"
	} else {
		shell = "/bin/sh"
		flag = "-c"
	}

	if err != nil {
		return fmt.Errorf("failed to initialize pipe for output: %s", err)
	}

	ctx := context.Background()
	if r.Timeout != 0 {
		ctx, _ = context.WithTimeout(ctx, r.Timeout)
	}

	// Setup the command
	cmd := exec.CommandContext(ctx, shell, flag, r.Inline)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	log.Printf("[INFO] Executing: %s %s \"%s\"",
		shell, flag, r.Inline)

	// Start the command
	err = cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if err != nil {
		msg := fmt.Errorf("[ERROR] Error running command '%s': %v",
			r.Inline, err)
		if r.IgnoreErrors {
			log.Print(msg)
			err = nil
			return
		}
		err = msg
	}
	return
}
