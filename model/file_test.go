package model

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"strings"
	"testing"
	"time"
)

func TestInitRedisClient(t *testing.T) {
	fc := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	// set url
	req.Header.SetMethod("GET")
	url := "https://apipost.oss-cn-beijing.aliyuncs.com/kunpeng/test/c35a24da-6958-4e74-b9e5-dc6f081dae3d.txt"
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()

	if err := fc.Do(req, resp); err != nil {
		fmt.Println("请求错误", err)
	}
	strs := strings.Split(string(resp.Body()), "\n")
	for _, str := range strs {
		fmt.Println("str:            ", str)
		time.Sleep(5 * time.Second)
	}

}

func TestParameterizedFile_UseFile(t *testing.T) {
	data := "data\n\"{\"\"data\"\": {\"\"efficacy_phrase\"\": [\"\"促进内渗\"\", \"\"有效剪切力\"\", \"\"防止起泡\"\"], \"\"lang\"\": \"\"cn\"\", \"\"raw_query\"\": \"\"促进内渗有效剪切力防止起泡\"\", \"\"patent_ids\"\": [\"\"b87cd577-2397-4a44-96cd-c544e3145974\"\"]}}\"\n\"{\"\"data\"\": {\"\"efficacy_phrase\"\": [\"\"保持色谱效率\"\", \"\"稳定误差信号\"\"], \"\"lang\"\": \"\"cn\"\", \"\"raw_query\"\": \"\"保持色谱效率稳定误差信号\"\", \"\"patent_ids\"\": [\"\"e87f17c0-d4b4-4aa0-8fd3-9fe2a33d992f\"\", \"\"782d3de2-04d8-4861-9e9b-764ed7024f24\"\", \"\"cb6908d1-9f70-407e-849f-6f0334107a26\"\", \"\"f6856fcc-176b-40f7-b902-8e09fda58c50\"\", \"\"fd9809ef-fc69-4821-b67e-6a1d2014922c\"\"]}}\"\n\"{\"\"data\"\": {\"\"efficacy_phrase\"\": [\"\"剃须性能得到提升\"\", \"\"易于免疫沉淀\"\", \"\"改善均匀性和平坦度\"\"], \"\"lang\"\": \"\"cn\"\", \"\"raw_query\"\": \"\"剃须性能得到提升易于免疫沉淀改善均匀性和平坦度\"\", \"\"patent_ids\"\": []}}\"\n\"{\"\"data\"\": {\"\"efficacy_phrase\"\": [\"\"点长最小化\"\", \"\"节省维修费用\"\"], \"\"lang\"\": \"\"cn\"\", \"\"raw_query\"\": \"\"点长最小化节省维修费用\"\", \"\"patent_ids\"\": []}}\"\n\"{\"\"data\"\": {\"\"efficacy_phrase\"\": [], \"\"lang\"\": \"\"cn\"\", \"\"raw_query\"\": \"\"高信号功率电平\"\", \"\"patent_ids\"\": [\"\"73bfa0c7-77b2-4315-9491-8679f46e61af\"\", \"\"30e36364-5d14-4011-99f8-1686bf9dadda\"\", \"\"d4c86ef6-ca66-47a2-b9c4-1fefa0eec124\"\", \"\"0ccec102-6dc5-4a93-8592-226a95d202b5\"\", \"\"ca5914a3-f089-4f30-bf9d-f5d6fdb73640\"\"]}}\"\n\"{\"\"data\"\": {\"\"efficacy_phrase\"\": [], \"\"lang\"\": \"\"cn\"\", \"\"raw_query\"\": \"\"增加整体药物释放\"\", \"\"patent_ids\"\": [\"\"c56b993e-ed42-476c-8a22-70ccfcc2480a\"\", \"\"0c7f6efe-327b-470e-9a8e-2f5974de2527\"\", \"\"233d8005-e2e7-46c1-a241-ae2c924a4473\"\", \"\"6ee72205-aa9f-425a-b2aa-3c171e98d81d\"\", \"\"4b9fe50f-b378-43f8-855b-bdb57fd44daa\"\"]}}\"\n\"{\"\"data\"\": {\"\"efficacy_phrase\"\": [\"\"等延迟\"\"], \"\"lang\"\": \"\"cn\"\", \"\"raw_query\"\": \"\"等延迟\"\", \"\"patent_ids\"\": [\"\"1254e8ad-5699-4c6c-954f-6780d822b23f\"\", \"\"0a63c921-1cb8-4a83-a862-1b9fb4053a54\"\", \"\"c8798788-aef3-4ab9-ab23-5490b195069d\"\", \"\"89d69caf-f54c-460d-94f7-3e595cef2914\"\", \"\"cb99d5e5-9c8f-409c-88ce-d206982ddc84\"\"]}}\"\n\"{\"\"data\"\": {\"\"efficacy_phrase\"\": [\"\"第一个成员 1 和第二个成员 2 之间的拟合精度\"\", \"\"复性产量的提高\"\", \"\"透光性好\"\"], \"\"lang\"\": \"\"cn\"\", \"\"raw_query\"\": \"\"第一个成员 1 和第二个成员 2 之间的拟合精度复性产量的提高透光性好\"\", \"\"patent_ids\"\": []}}\"\n\"{\"\"data\"\": {\"\"efficacy_phrase\"\": [\"\"便于轮换\"\", \"\"行动中的控制力大大增强\"\", \"\"提高亲水性和乳化稳定性\"\"], \"\"lang\"\": \"\"cn\"\", \"\"raw_query\"\": \"\"便于轮换行动中的控制力大大增强提高亲水性和乳化稳定性\"\", \"\"patent_ids\"\": []}}\"\n\"{\"\"data\"\": {\"\"efficacy_phrase\"\": [\"\"减少间隙尺寸\"\", \"\"维护结构\"\"], \"\"lang\"\": \"\"cn\"\", \"\"raw_query\"\": \"\"减少间隙尺寸维护结构\"\", \"\"patent_ids\"\": [\"\"6e506480-6c46-40dc-be8d-ed5150da9b08\"\", \"\"4f4f740b-14d3-4057-8b3e-bde78acf5087\"\", \"\"afaae0cf-8cdf-4cfc-aad0-0a1c472ed347\"\", \"\"9ad3242d-c385-4ebb-bdaf-bec78c0a3bbc\"\", \"\"d12288de-1a7f-483f-8bea-fde6d15fa7ba\"\"]}}\"\n\"{\"\"data\"\": {\"\"efficacy_phrase\"\": [], \"\"lang\"\": \"\"cn\"\", \"\"raw_query\"\": \"\"轻松顺畅地旅行优异的传导稳定性\"\", \"\"patent_ids\"\": []}}\"\n\"{\"\"data\"\": {\"\"efficacy_phrase\"\": [\"\"组装能力高\"\"], \"\"lang\"\": \"\"cn\"\", \"\"raw_query\"\": \"\"组装能力高\"\", \"\"patent_ids\"\": [\"\"d595c7f6-cd50-4028-90fe-b0b57f682fd5\"\", \"\"d8bb6879-bc25-4818-84da-7152cf64e5d7\"\", \"\"69fc48b9-5021-4b23-a563-d08e6847a0f5\"\", \"\"42e40137-776d-4399-9cfe-8e827d6bd29c\"\", \"\"42e927f6-f62b-412d-8251-5c4c35c65897\"\"]}}\"\n"
	// ParameterizedFile 参数化文件

	dataList := strings.Split(data, "\n")
	fmt.Println("data:      ", dataList)
	for i := 0; i < len(dataList); i++ {

	}

}
