package lib

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/mbndr/logo"
	"os"
	"testing"
)

func TestNoDefaultProfile(t *testing.T) {
	os.Setenv("AWS_CONFIG_FILE", "test/no_content.cfg")
	p, err := defaultProfile()
	if err == nil {
		t.Errorf("Expected error with no default section, but received: %+v", p)
	} else {
		t.Logf("Expected no default profile error: %v", err)
	}
}

func TestEmptyDefaultProfile(t *testing.T) {
	os.Setenv("AWS_CONFIG_FILE", "test/empty_default.cfg")
	p, err := defaultProfile()
	if err != nil {
		t.Errorf("Error getting default profile: %v", err)
	}

	if p.name != "default" {
		t.Errorf("Unexpected profile name, expected: default, got: %s", p.name)
	}
}

func ExampleDefaultProfile() {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")
	p, err := defaultProfile()
	if err != nil {
		fmt.Printf("Error getting default profile: %v/n", err)
	}

	if p != nil {
		fmt.Println(p.name)
		fmt.Println(p.Region)
	}
	// Output:
	// default
	// us-east-1
}

func ExampleDefaultProfileEnv() {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")
	os.Setenv("AWS_DEFAULT_PROFILE", "alt_default")

	p, err := defaultProfile()
	if err != nil {
		fmt.Printf("Error getting default profile: %v/n", err)
	}

	// Unset, otherwise it's carried through to remaining tests
	os.Unsetenv("AWS_DEFAULT_PROFILE")

	if p != nil {
		fmt.Println(p.name)
		fmt.Println(p.Region)
		fmt.Println(p.MfaSerial)
	}
	// Output:
	// alt_default
	// us-west-1
	// 12345678
}

func TestNilProfileName(t *testing.T) {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")

	p, err := getProfile(nil)
	if err == nil {
		t.Errorf("Expected error with nil profile name, but received: %+v", p)
	} else {
		t.Logf("Expected nil profile name error: %v", err)
	}
}

func TestEmptyProfileName(t *testing.T) {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")

	p, err := getProfile(aws.String(""))
	if err == nil {
		t.Errorf("Expected error for empty profile name, but received: %+v", p)
	} else {
		t.Logf("Expected empty profile name error: %v", err)
	}
}

func TestBadProfileName(t *testing.T) {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")
	name := "lkajiq"

	p, err := getProfile(aws.String(name))
	if err == nil {
		t.Errorf("Expected error for bad profile name, but received: %+v", p)
	} else {
		t.Logf("Expected bad profile name error: %v", err)
	}
}

func TestUnparsableRoleArn(t *testing.T) {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")

	p, err := getProfile(aws.String("has_bad_role"))
	if err == nil {
		t.Errorf("Expected error for unparsable role arn, but received: %+v", p)
	} else {
		t.Logf("Expected unparsable arn error: %v", err)
	}
}

func TestNonIamRoleArn(t *testing.T) {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")

	p, err := getProfile(aws.String("has_s3_arn"))
	if err == nil {
		t.Errorf("Expected error for non-iam role arn, but received: %+v", p)
	} else {
		t.Logf("Expected non-iam arn error: %v", err)
	}
}

func TestBadSourceProfile(t *testing.T) {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")
	name := "has_role_bad_source"

	// We're not validating if source_profile is a valid name, should we?
	// or leave it up to the SDK to handle that?
	p, err := getProfile(aws.String(name))
	if err != nil {
		t.Errorf("Error looking up profile %s: %v", name, err)
	}

	if p.SourceProfile != "bad" && p.name != name {
		t.Errorf("Bad profile returned, expecting source_profile = 'bad' but got: %s", p.SourceProfile)
	}
}

func TestNoSourceProfile(t *testing.T) {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")
	name := "has_role_no_source"

	p, err := getProfile(aws.String(name))
	if err != nil {
		t.Errorf("Error looking up profile %s: %v", name, err)
	}

	if p.SourceProfile != "default" && p.name != name {
		t.Errorf("Bad profile returned, expecting default source_profile but got: %s", p.SourceProfile)
	}
}

func ExampleGetProfileNoRoleMfa() {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")
	name := "basic"

	p, err := getProfile(aws.String(name))
	if err != nil {
		fmt.Printf("Error getting profile %s: %v", name, err)
	}

	fmt.Println(p.name)
	fmt.Println(p.SourceProfile)
	fmt.Println(p.RoleArn)
	fmt.Println(p.MfaSerial)
	// Output:
	// basic
	// default
	//
	//
}

func ExampleGetProfileRoleNoMfa() {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")
	name := "has_role"

	p, err := getProfile(aws.String(name))
	if err != nil {
		fmt.Printf("Error getting profile %s: %v", name, err)
	}

	fmt.Println(p.name)
	fmt.Println(p.SourceProfile)
	fmt.Println(p.RoleArn)
	fmt.Println(p.MfaSerial)
	// Output:
	// has_role
	// default
	// arn:aws:iam::012345678901:mfa/my_iam_user
	//
}

func ExampleGetProfileRoleMfa() {
	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")
	name := "has_role_explicit_mfa"

	p, err := getProfile(aws.String(name))
	if err != nil {
		fmt.Printf("Error getting profile %s: %v", name, err)
	}

	fmt.Println(p.name)
	fmt.Println(p.SourceProfile)
	fmt.Println(p.RoleArn)
	fmt.Println(p.MfaSerial)
	// Output:
	// has_role_explicit_mfa
	// default
	// arn:aws:iam::012345678901:mfa/my_iam_user
	// 87654321
}

// FIXME fails because we're not resolving source_profile after getting named profile
// should this be a feature of this code, since it breaks aws sdk compatibility?
//func ExampleGetProfileRoleInheritMfa() {
//	os.Setenv("AWS_CONFIG_FILE", "test/aws.cfg")
//	name := "has_role_inherit_mfa"
//

//	p, err := getProfile(aws.String(name))
//	if err != nil {
//		fmt.Printf("Error getting profile %s: %v", name, err)
//	}
//
//	fmt.Println(p.name)
//	fmt.Println(p.SourceProfile)
//	fmt.Println(p.RoleArn)
//	fmt.Println(p.MfaSerial)
//	// Output:
//	// has_role_inherit_mfa
//	// alt_default
//	// arn:aws:iam::012345678901:mfa/my_iam_user
//	// 12345678
//}

func defaultProfile() (*AWSProfile, error) {
	cm, err := NewAwsConfigManager(logo.INFO)
	if err != nil {
		return nil, err
	}

	p, err := cm.DefaultProfile()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func getProfile(name *string) (*AWSProfile, error) {
	cm, err := NewAwsConfigManager(logo.INFO)
	if err != nil {
		return nil, err
	}

	p, err := cm.GetProfile(name)
	if err != nil {
		return nil, err
	}

	return p, nil
}
