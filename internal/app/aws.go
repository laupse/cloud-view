package app

import (
	"context"
	"fmt"
	// "strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	clog "github.com/charmbracelet/log"
	"github.com/sourcegraph/conc/pool"
)

type AwsVpcRequest struct {
	InternetGateway *string `query:"internet_gateway,omitempty"`
	Subnet          *string `query:"subnet,omitempty"`
	RouteTables     *string `query:"route_tables,omitempty"`
	Instances       *string `query:"instances,omitempty"`
}

func (filter AwsVpcRequest) ValueAndTag() {

}

type AwsIconUrl string

type AwsGraphProvider struct {
	Ec2Client *ec2.Client

	Request *AwsVpcRequest
}

var (
	vpcIcon      AwsIconUrl = "https://icons.terrastruct.com/aws%2FNetworking%20&%20Content%20Delivery%2FAmazon-VPC.svg"
	subnetIcon   AwsIconUrl = "https://icons.terrastruct.com/aws%2F_Group%20Icons%2FVPC-subnet-private_light-bg.svg"
	igwIcon      AwsIconUrl = "https://icons.terrastruct.com/aws%2FNetworking%20&%20Content%20Delivery%2FAmazon-VPC_Internet-Gateway_light-bg.svg"
	natgwIcon    AwsIconUrl = "https://icons.terrastruct.com/aws%2FNetworking%20&%20Content%20Delivery%2FAmazon-VPC_NAT-Gateway_light-bg.svg"
	rtIcon       AwsIconUrl = "https://icons.terrastruct.com/aws%2FNetworking%20&%20Content%20Delivery%2FAmazon-Route-53_Route-Table_light-bg.svg"
	instanceIcon AwsIconUrl = "https://icons.terrastruct.com/aws%2FCompute%2F_Instance%2FAmazon-EC2_Instance_light-bg.svg"
)

func (provider *AwsGraphProvider) Create(ctx context.Context, graphUpdateChannel chan<- graphUpdate) error {
	fetchPool := pool.New().WithContext(ctx)
	fetchPool.Go(func(context context.Context) error {
		clog.Info("fetching vpc")
		err := provider.insertAwsVPC(ctx, graphUpdateChannel)
		if err != nil {
			return err
		}
		return nil
	})
	if provider.Request.Instances != nil {
		fetchPool.Go(func(context context.Context) error {
			clog.Info("fetching instance")
			err := provider.insertEc2Instances(ctx, graphUpdateChannel)
			if err != nil {
				return err
			}
			return nil
		})
	}

	if provider.Request.Subnet != nil {
		fetchPool.Go(func(context context.Context) error {
			clog.Info("fetching subnet")
			err := provider.insertSubnets(ctx, graphUpdateChannel)
			if err != nil {
				return err
			}
			return nil
		})
	}

	if provider.Request.InternetGateway != nil {
		fetchPool.Go(func(context context.Context) error {
			clog.Info("fetching internet gateway")
			err := provider.insertInternetGateway(ctx, graphUpdateChannel)
			if err != nil {
				return err
			}
			return nil
		})
	}

	if provider.Request.RouteTables != nil {
		fetchPool.Go(func(context context.Context) error {
			clog.Info("fetching route table")
			err := provider.insertRouteTables(ctx, graphUpdateChannel)
			if err != nil {
				return err
			}
			return nil
		})
	}

	err := fetchPool.Wait()
	if err != nil {
		clog.Error("error encountered when fetching from aws %s", err)
	}

	return nil
}

func (provider *AwsGraphProvider) insertAwsVPC(context context.Context, update chan<- graphUpdate) error {
	vpcsInputs := ec2.DescribeVpcsInput{}
	vpcsOutputs, err := provider.Ec2Client.DescribeVpcs(context, &vpcsInputs)
	if err != nil {
		return err
	}
	for _, vpc := range vpcsOutputs.Vpcs {
		update <- graphUpdate{action: Create, key: *vpc.VpcId}
		update <- graphUpdate{action: Set, key: fmt.Sprintf("%s.icon", *vpc.VpcId), value: string(vpcIcon)}
	}

	return nil
}

func (provider *AwsGraphProvider) insertSubnets(context context.Context, update chan<- graphUpdate) error {
	subnetsInputs := ec2.DescribeSubnetsInput{}
	subnetsOutputs, err := provider.Ec2Client.DescribeSubnets(context, &subnetsInputs)
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

func (provider *AwsGraphProvider) insertEc2Instances(context context.Context, update chan<- graphUpdate) error {
	instancesInputs := ec2.DescribeInstancesInput{}
	instancesOutputs, err := provider.Ec2Client.DescribeInstances(context, &instancesInputs)
	if err != nil {
		return err
	}
	for _, reservation := range instancesOutputs.Reservations {
		for _, instance := range reservation.Instances {
			shape := fmt.Sprintf("%s.%s.%s", *instance.VpcId, *instance.SubnetId, *instance.InstanceId)
			// instanceType := strings.ToUpper(strings.Split(string(instance.InstanceType), ".")[0])
			// icon := "https://icons.terrastruct.com/aws%2FCompute%2F_Instance%2FAmazon-EC2_" + instanceType + "-Instance_light-bg.svg"
			update <- graphUpdate{action: Create, key: shape}
			update <- graphUpdate{action: Set, key: fmt.Sprintf("%s.icon", shape), value: string(instanceIcon)}
		}
	}

	return nil
}

func (provider *AwsGraphProvider) insertInternetGateway(context context.Context, update chan<- graphUpdate) error {
	internetGatewaysInputs := ec2.DescribeInternetGatewaysInput{}
	internetGatewaysOutputs, err := provider.Ec2Client.DescribeInternetGateways(context, &internetGatewaysInputs)
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

func (provider *AwsGraphProvider) insertNatGateway(context context.Context, update chan<- graphUpdate) error {
	natGatewaysInputs := ec2.DescribeNatGatewaysInput{}
	natGatewaysOutputs, err := provider.Ec2Client.DescribeNatGateways(context, &natGatewaysInputs)
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

func (provider *AwsGraphProvider) insertRouteTables(context context.Context, update chan<- graphUpdate) error {
	routeTablesInput := ec2.DescribeRouteTablesInput{}
	routeTablesOutput, err := provider.Ec2Client.DescribeRouteTables(context, &routeTablesInput)
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
