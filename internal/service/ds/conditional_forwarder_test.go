package ds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
)

func TestAccDirectoryServiceConditionalForwarder_Condition_basic(t *testing.T) {
	resourceName := "aws_directory_service_conditional_forwarder.fwd"

	ip1, ip2, ip3 := "8.8.8.8", "1.1.1.1", "8.8.4.4"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckDirectoryService(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConditionalForwarderDestroy,
		Steps: []resource.TestStep{
			// test create
			{
				Config: testAccDirectoryServiceConditionalForwarderConfig(ip1, ip2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionalForwarderExists(
						resourceName,
						[]string{ip1, ip2},
					),
				),
			},
			// test update
			{
				Config: testAccDirectoryServiceConditionalForwarderConfig(ip1, ip3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionalForwarderExists(
						resourceName,
						[]string{ip1, ip3},
					),
				),
			},
			// test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckConditionalForwarderDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DirectoryServiceConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_directory_service_conditional_forwarder" {
			continue
		}

		directoryId, domainName, err := tfds.ParseDSConditionalForwarderID(rs.Primary.ID)
		if err != nil {
			return err
		}

		res, err := conn.DescribeConditionalForwarders(&directoryservice.DescribeConditionalForwardersInput{
			DirectoryId:       aws.String(directoryId),
			RemoteDomainNames: []*string{aws.String(domainName)},
		})

		if tfawserr.ErrMessageContains(err, directoryservice.ErrCodeEntityDoesNotExistException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if len(res.ConditionalForwarders) > 0 {
			return fmt.Errorf("Expected AWS Directory Service Conditional Forwarder to be gone, but was still found")
		}
	}

	return nil
}

func testAccCheckConditionalForwarderExists(name string, dnsIps []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		directoryId, domainName, err := tfds.ParseDSConditionalForwarderID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectoryServiceConn

		res, err := conn.DescribeConditionalForwarders(&directoryservice.DescribeConditionalForwardersInput{
			DirectoryId:       aws.String(directoryId),
			RemoteDomainNames: []*string{aws.String(domainName)},
		})

		if err != nil {
			return err
		}

		if len(res.ConditionalForwarders) == 0 {
			return fmt.Errorf("No Conditional Fowrwarder found")
		}

		cfd := res.ConditionalForwarders[0]

		if dnsIps != nil {
			if len(dnsIps) != len(cfd.DnsIpAddrs) {
				return fmt.Errorf("DnsIpAddrs length mismatch")
			}

			for k, v := range cfd.DnsIpAddrs {
				if *v != dnsIps[k] {
					return fmt.Errorf("DnsIp mismatch, '%s' != '%s' at index '%d'", *v, dnsIps[k], k)
				}
			}
		}

		return nil
	}
}

func testAccDirectoryServiceConditionalForwarderConfig(ip1, ip2 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_directory_service_directory" "bar" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.main.id
    subnet_ids = [aws_subnet.foo.id, aws_subnet.bar.id]
  }

  tags = {
    Name = "terraform-testacc-directory-service-conditional-forwarder"
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-directory-service-conditional-forwarder"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = "terraform-testacc-directory-service-conditional-forwarder"
  }
}

resource "aws_subnet" "bar" {
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"

  tags = {
    Name = "terraform-testacc-directory-service-conditional-forwarder"
  }
}

resource "aws_directory_service_conditional_forwarder" "fwd" {
  directory_id = aws_directory_service_directory.bar.id

  remote_domain_name = "test.example.com"

  dns_ips = [
    "%s",
    "%s",
  ]
}
`, ip1, ip2)
}
