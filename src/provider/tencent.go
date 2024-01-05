package provider

import (
	"WHMCS-Chinese/src/utils"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tmt "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tmt/v20180321"
)

var TencentSecret = ""
var TencentID = ""

type Tencent struct {
	client *tmt.Client

	largeBatch int
}

func SetupTencent(largeBatch int) (*Tencent, error) {
	credential := common.NewCredential(
		TencentID,
		TencentSecret,
	)

	client, err := tmt.NewClient(credential, "ap-hongkong", profile.NewClientProfile())

	return &Tencent{
		client: client,
	}, err
}

func (t *Tencent) Translate(str []string, isLarge bool) ([]string, error) {
	keyword, strArr := PreProcess(str)

	request := tmt.NewTextTranslateRequest()

	request.SourceText = common.StringPtr(keyword)
	request.Source = common.StringPtr("en")
	request.Target = common.StringPtr("zh")
	request.ProjectId = common.Int64Ptr(0)

	var resp *tmt.TextTranslateResponse
	err := utils.BackOff(func() error {
		res, err := t.client.TextTranslate(request)
		if err != nil {
			return err
		}

		resp = res
		return nil
	})

	if err != nil {
		return nil, err
	}

	return PostProcess(strArr, *resp.Response.TargetText), nil
}
