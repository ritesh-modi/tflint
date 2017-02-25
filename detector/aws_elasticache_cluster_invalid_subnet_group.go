package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsElastiCacheClusterInvalidSubnetGroupDetector struct {
	*Detector
	IssueType         string
	Target            string
	DeepCheck         bool
	cacheSubnetGroups map[string]bool
}

func (d *Detector) CreateAwsElastiCacheClusterInvalidSubnetGroupDetector() *AwsElastiCacheClusterInvalidSubnetGroupDetector {
	return &AwsElastiCacheClusterInvalidSubnetGroupDetector{
		Detector:          d,
		IssueType:         issue.ERROR,
		Target:            "aws_elasticache_cluster",
		DeepCheck:         true,
		cacheSubnetGroups: map[string]bool{},
	}
}

func (d *AwsElastiCacheClusterInvalidSubnetGroupDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeCacheSubnetGroups()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, subnetGroup := range resp.CacheSubnetGroups {
		d.cacheSubnetGroups[*subnetGroup.CacheSubnetGroupName] = true
	}
}

func (d *AwsElastiCacheClusterInvalidSubnetGroupDetector) Detect(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_elasticache_cluster").Items {
			subnetGroupToken, err := hclLiteralToken(item, "subnet_group_name")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			subnetGroup, err := d.evalToString(subnetGroupToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if !d.cacheSubnetGroups[subnetGroup] {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is invalid subnet group name.", subnetGroup),
					Line:    subnetGroupToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
