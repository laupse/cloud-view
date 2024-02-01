package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AwsVpcRequest struct {
	InternetGateway *string `json:"internet_gateway,omitempty"`
	Subnet          *string `json:"subnet,omitempty"`
	RouteTables     *string `json:"route_tables,omitempty"`
	Instances       *string `json:"instances,omitempty"`
}

type AwsIconUrl string

var (
	vpcIcon    AwsIconUrl = "https://icons.terrastruct.com/aws%2FNetworking%20&%20Content%20Delivery%2FAmazon-VPC.svg"
	subnetIcon AwsIconUrl = "https://icons.terrastruct.com/aws%2F_Group%20Icons%2FVPC-subnet-private_light-bg.svg"
	igwIcon    AwsIconUrl = "https://icons.terrastruct.com/aws%2FNetworking%20&%20Content%20Delivery%2FAmazon-VPC_Internet-Gateway_light-bg.svg"
	natgwIcon  AwsIconUrl = "https://icons.terrastruct.com/aws%2FNetworking%20&%20Content%20Delivery%2FAmazon-VPC_NAT-Gateway_light-bg.svg"
	rtIcon     AwsIconUrl = "https://icons.terrastruct.com/aws%2FNetworking%20&%20Content%20Delivery%2FAmazon-Route-53_Route-Table_light-bg.svg"
)

func (app *App) InsertAwsVPC(context context.Context, update chan<- graphUpdate) error {
	vpcsInputs := ec2.DescribeVpcsInput{}
	vpcsOutputs, err := app.Ec2Client.DescribeVpcs(context, &vpcsInputs)
	if err != nil {
		return err
	}
	for _, vpc := range vpcsOutputs.Vpcs {
		update <- graphUpdate{action: Create, key: *vpc.VpcId}
		update <- graphUpdate{action: Set, key: fmt.Sprintf("%s.icon", *vpc.VpcId), value: string(vpcIcon)}
	}

	return nil
}

func (app *App) InsertSubnets(context context.Context, update chan<- graphUpdate) error {
	subnetsInputs := ec2.DescribeSubnetsInput{}
	subnetsOutputs, err := app.Ec2Client.DescribeSubnets(context, &subnetsInputs)
	if err != nil {
		return err
	}
	for _, subnets := range subnetsOutputs.Subnets {
		shape := fmt.Sprintf("%s.%s", *subnets.VpcId, *subnets.SubnetId)
		update <- graphUpdate{action: Create, key: shape}
		update <- graphUpdate{action: Set, key: fmt.Sprintf("%s.icon", shape), value: string(subnetIcon)}
	}

	return nil
}

func (app *App) InsertEc2Instances(context context.Context, update chan<- graphUpdate) error {
	instancesInputs := ec2.DescribeInstancesInput{}
	instancesOutputs, err := app.Ec2Client.DescribeInstances(context, &instancesInputs)
	if err != nil {
		return err
	}
	for _, reservation := range instancesOutputs.Reservations {
		for _, instance := range reservation.Instances {
			shape := fmt.Sprintf("%s.%s.%s", *instance.VpcId, *instance.SubnetId, *instance.InstanceId)
			instanceType := strings.ToUpper(strings.Split(string(instance.InstanceType), ".")[0])
			icon := "https://icons.terrastruct.com/aws%2FCompute%2F_Instance%2FAmazon-EC2_" + instanceType + "-Instance_light-bg.svg"
			update <- graphUpdate{action: Create, key: shape}
			update <- graphUpdate{action: Set, key: fmt.Sprintf("%s.icon", shape), value: icon}
		}
	}

	return nil
}

func (app *App) InsertInternetGateway(context context.Context, update chan<- graphUpdate) error {
	internetGatewaysInputs := ec2.DescribeInternetGatewaysInput{}
	internetGatewaysOutputs, err := app.Ec2Client.DescribeInternetGateways(context, &internetGatewaysInputs)
	if err != nil {
		return err
	}

	for _, internetGateways := range internetGatewaysOutputs.InternetGateways {
		for _, attachment := range internetGateways.Attachments {
			shape := fmt.Sprintf("%s.%s", *attachment.VpcId, *internetGateways.InternetGatewayId)
			update <- graphUpdate{action: Create, key: shape}
			update <- graphUpdate{action: Set, key: fmt.Sprintf("%s.icon", shape), value: string(igwIcon)}
			update <- graphUpdate{action: Set, key: fmt.Sprintf("%s.shape", shape), value: "image"}
		}
		if err != nil {
			return err
		}
	}

	if len(internetGatewaysOutputs.InternetGateways) > 0 {
		update <- graphUpdate{action: Set, key: "internet.icon", value: "https://icons.terrastruct.com/aws%2F_General%2FInternet-alt2_light-bg.svg"}
		update <- graphUpdate{action: Set, key: "internet.shape", value: "image"}
	}
	return nil

}

func (app *App) InsertNatGateway(context context.Context, update chan<- graphUpdate) error {
	natGatewaysInputs := ec2.DescribeNatGatewaysInput{}
	natGatewaysOutputs, err := app.Ec2Client.DescribeNatGateways(context, &natGatewaysInputs)
	if err != nil {
		return err
	}

	for _, natGateway := range natGatewaysOutputs.NatGateways {
		shape := fmt.Sprintf("%s.%s.%s", *natGateway.VpcId, *natGateway.SubnetId, *natGateway.NatGatewayId)
		update <- graphUpdate{action: Create, key: shape}
		update <- graphUpdate{action: Set, key: fmt.Sprintf("%s.icon", shape), value: string(natgwIcon)}
		update <- graphUpdate{action: Set, key: fmt.Sprintf("%s.shape", shape), value: "image"}
	}
	return nil
}

func (app *App) InsertRouteTables(context context.Context, update chan<- graphUpdate) error {
	routeTablesInput := ec2.DescribeRouteTablesInput{}
	routeTablesOutput, err := app.Ec2Client.DescribeRouteTables(context, &routeTablesInput)
	if err != nil {
		return err
	}

	for _, routeTable := range routeTablesOutput.RouteTables {
		shape := fmt.Sprintf("%s.%s", *routeTable.VpcId, *routeTable.RouteTableId)
		update <- graphUpdate{action: Create, key: shape}
		update <- graphUpdate{action: Set, key: fmt.Sprintf("%s.icon", shape), value: string(rtIcon)}
		update <- graphUpdate{action: Set, key: fmt.Sprintf("%s.shape", shape), value: "image"}
		if err != nil {
			return err
		}
		for _, associations := range routeTable.Associations {
			if associations.SubnetId != nil {
				shape := fmt.Sprintf("%s.%s <-> %s.%s", *routeTable.VpcId, *routeTable.RouteTableId, *routeTable.VpcId, *associations.SubnetId)
				update <- graphUpdate{action: Create, key: shape}
			}
		}
	}
	return nil
}
