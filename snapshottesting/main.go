package snapshottesting

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
)

var refreshSnapshots = flag.Bool("refresh-snapshots", false, "For refresh snapshots add -refresh-snapshots flag")

func Init() {
	_, filename, _, _ := runtime.Caller(1)
	dir := path.Join(path.Dir(filename), "../..")
	fmt.Println(dir)
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func Run(t *testing.T, stacks []awscdk.Stack) {
	t.Log("For refresh snapshots add -refresh-snapshots flag")
	for _, stack := range stacks {
		template := assertions.Template_FromStack(
			stack,
			&assertions.TemplateParsingOptions{},
		)
		jsonTemplate, err := json.Marshal(
			template.ToJSON(),
		)
		if err != nil {
			log.Fatalf("Parsing templates error:%s", err)
		}
		if *refreshSnapshots {
			err = os.WriteFile(
				fmt.Sprintf("snapshots/%s.json", *stack.StackName()),
				jsonTemplate,
				0600,
			)
		}
		if err != nil {
			log.Fatalf("Writing files error:%s", err)
		}
		ret, err := os.Open(fmt.Sprintf("snapshots/%s.json", *stack.StackName()))
		if err != nil {
			t.Fatalf("Not exist snapshot: %s", *stack.StackName())
		}
		b, _ := io.ReadAll(ret)
		jsonString := string(b)
		snapshotTemplate := assertions.Template_FromString(
			jsii.String(jsonString),
			&assertions.TemplateParsingOptions{},
		)
		defer func() {
			if err := recover(); err != nil {
				if !strings.Contains(fmt.Sprint(err), "/Properties/Code/S3Key") {
					t.Fatal(err)
				}
			}
		}()
		template.TemplateMatches(snapshotTemplate.ToJSON())
	}
}
