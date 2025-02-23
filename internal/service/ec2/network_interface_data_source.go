package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceNetworkInterface() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNetworkInterfaceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
			"association": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"association_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"carrier_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"customer_owned_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip_owner_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_dns_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"attachment": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attachment_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"instance_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_owner_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"interface_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"mac_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"requester_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceNetworkInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeNetworkInterfacesInput{}
	if v, ok := d.GetOk("id"); ok {
		input.NetworkInterfaceIds = []*string{aws.String(v.(string))}
	}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = BuildFiltersDataSource(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Reading Network Interface: %s", input)
	resp, err := conn.DescribeNetworkInterfaces(input)
	if err != nil {
		return err
	}

	if resp == nil || len(resp.NetworkInterfaces) == 0 {
		return fmt.Errorf("no matching network interface found")
	}

	if len(resp.NetworkInterfaces) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more specific search criteria")
	}

	eni := resp.NetworkInterfaces[0]

	d.SetId(aws.StringValue(eni.NetworkInterfaceId))
	if eni.Association != nil {
		d.Set("association", flattenNetworkInterfaceAssociation(eni.Association))
	}
	if eni.Attachment != nil {
		attachment := []interface{}{FlattenAttachment(eni.Attachment)}
		d.Set("attachment", attachment)
	}
	d.Set("availability_zone", eni.AvailabilityZone)
	d.Set("description", eni.Description)
	d.Set("security_groups", FlattenGroupIdentifiers(eni.Groups))
	d.Set("interface_type", eni.InterfaceType)
	d.Set("ipv6_addresses", flattenNetworkInterfaceIPv6Address(eni.Ipv6Addresses))
	d.Set("mac_address", eni.MacAddress)
	d.Set("owner_id", eni.OwnerId)
	d.Set("private_dns_name", eni.PrivateDnsName)
	d.Set("private_ip", eni.PrivateIpAddress)
	d.Set("private_ips", FlattenNetworkInterfacesPrivateIPAddresses(eni.PrivateIpAddresses))
	d.Set("requester_id", eni.RequesterId)
	d.Set("subnet_id", eni.SubnetId)
	d.Set("outpost_arn", eni.OutpostArn)
	d.Set("vpc_id", eni.VpcId)

	if err := d.Set("tags", KeyValueTags(eni.TagSet).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
