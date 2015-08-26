package main

import (
	"io/ioutil"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/k0kubun/pp"
	"github.com/mob-manzoku/go-jsonp"
	"gopkg.in/yaml.v2"
)

type instanceTypes map[string]*size

type size struct {
	size           *string
	vCPU           *int
	ECU            *string
	memoryGiB      *float64
	storageGB      *string
	system         *string
	piopsOptimized *bool
	price          *float64
	network        *string
}

var awsPricingURLs = map[string]string{
	"ec2":         "http://a0.awsstatic.com/pricing/1/ec2/linux-od.min.js",
	"rds":         "http://a0.awsstatic.com/pricing/1/rds/mysql/pricing-standard-deployments.min.js",
	"elasticache": "http://a0.awsstatic.com/pricing/1/elasticache/pricing-standard-deployments-elasticache.min.js",
}

func main() {
	//region := "apac-tokyo"

	ec2 := GetEC2Pricing("ap-northeast-1")
	ec := GetElasticachePricing("ap-northeast-1")
	rds := GetRDSPricing("apac-tokyo")
	pp.Print(ec2)
	pp.Print(ec)
	pp.Print(rds)

}

func GetEC2Pricing(region string) instanceTypes {

	service := "ec2"
	ret := instanceTypes{}

	str, _ := gojsonp.GetJSONFromURL(awsPricingURLs[service])

	raw, err := simplejson.NewJson([]byte(str))
	if err != nil {
		panic(err)
	}

	regions := raw.Get("config").Get("regions")

	for i := range regions.MustArray() {

		if regions.GetIndex(i).Get("region").MustString() != region {
			continue
		}

		types := regions.GetIndex(i).Get("instanceTypes")

		for j := range types.MustArray() {

			sizes := types.GetIndex(j).Get("sizes")
			for k := range sizes.MustArray() {
				s := sizes.GetIndex(k)

				msize := s.Get("size").MustString()
				mvCPU, _ := strconv.Atoi(s.Get("vCPU").MustString())
				mECU := s.Get("ECU").MustString()
				mmem, _ := strconv.ParseFloat(s.Get("memoryGiB").MustString(), 64)
				msto := s.Get("storageGB").MustString()
				msys := s.Get("valueColumns").GetIndex(0).Get("name").MustString()
				mprice, _ := strconv.ParseFloat(s.Get("valueColumns").GetIndex(0).Get("prices").Get("USD").MustString(), 64)

				obj := &size{
					size:      &msize,
					vCPU:      &mvCPU,
					ECU:       &mECU,
					memoryGiB: &mmem,
					storageGB: &msto,
					system:    &msys,
					price:     &mprice,
				}

				ret[*obj.size] = obj

			}
		}

	}
	return ret

}

func GetElasticachePricing(region string) instanceTypes {

	service := "elasticache"
	specpath := "spec/elasticache.yml"
	ret := instanceTypes{}

	str, _ := gojsonp.GetJSONFromURL(awsPricingURLs[service])

	raw, err := simplejson.NewJson([]byte(str))
	if err != nil {
		panic(err)
	}

	regions := raw.Get("config").Get("regions")

	for i := range regions.MustArray() {

		if regions.GetIndex(i).Get("region").MustString() != region {
			continue
		}

		types := regions.GetIndex(i).Get("types")

		for j := range types.MustArray() {

			tiers := types.GetIndex(j).Get("tiers")

			for k := range tiers.MustArray() {

				msize := tiers.GetIndex(k).Get("name").MustString()
				mprice, _ := strconv.ParseFloat(tiers.GetIndex(k).Get("prices").Get("USD").MustString(), 64)
				obj := &size{
					size:  &msize,
					price: &mprice,
				}

				ret[*obj.size] = obj
			}

		}
	}

	buf, err := ioutil.ReadFile(specpath)
	if err != nil {
		panic(err)
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &m)

	for _, v := range m[service].([]interface{}) {

		msize := v.(map[interface{}]interface{})["size"].(string)
		mvCPU := v.(map[interface{}]interface{})["vCPU"].(int)
		mmem := v.(map[interface{}]interface{})["memoryGiB"].(float64)
		mnet := v.(map[interface{}]interface{})["network"].(string)

		if _, ok := ret[msize]; ok {
			ret[msize].vCPU = &mvCPU
			ret[msize].memoryGiB = &mmem
			ret[msize].network = &mnet
		}

	}

	return ret
}

func GetRDSPricing(region string) instanceTypes {

	service := "rds"
	specpath := "spec/rds.yml"
	ret := instanceTypes{}

	str, _ := gojsonp.GetJSONFromURL(awsPricingURLs[service])

	raw, err := simplejson.NewJson([]byte(str))
	if err != nil {
		panic(err)
	}

	regions := raw.Get("config").Get("regions")

	for i := range regions.MustArray() {

		if regions.GetIndex(i).Get("region").MustString() != region {
			continue
		}

		types := regions.GetIndex(i).Get("types")

		for j := range types.MustArray() {

			tiers := types.GetIndex(j).Get("tiers")

			for k := range tiers.MustArray() {

				msize := tiers.GetIndex(k).Get("name").MustString()
				mprice, _ := strconv.ParseFloat(tiers.GetIndex(k).Get("prices").Get("USD").MustString(), 64)
				obj := &size{
					size:  &msize,
					price: &mprice,
				}

				ret[*obj.size] = obj
			}

		}
	}

	buf, err := ioutil.ReadFile(specpath)
	if err != nil {
		panic(err)
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &m)

	for _, v := range m[service].([]interface{}) {

		msize := v.(map[interface{}]interface{})["size"].(string)
		mvCPU := v.(map[interface{}]interface{})["vCPU"].(int)
		mpiop := v.(map[interface{}]interface{})["piopsOptimized"].(bool)
		mmem := v.(map[interface{}]interface{})["memoryGiB"].(float64)
		mnet := v.(map[interface{}]interface{})["network"].(string)

		if _, ok := ret[msize]; ok {
			ret[msize].vCPU = &mvCPU
			ret[msize].memoryGiB = &mmem
			ret[msize].piopsOptimized = &mpiop
			ret[msize].network = &mnet
		}

	}

	return ret
}
