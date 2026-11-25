// GSNewsPaper mirrors com.openttd.gamescript.GSNewsPaper: the one gamescript
// request modelled as its own class, the createNews command.
//
// Java's companyId is a short, -1 meaning "send to every company".

package gamescript

import "strings"

import "com.openttd/network/constant"

// NewsType: the news types a script can create news for. See news_type.h and
// script_news.hpp; this is the arbitrary selection of them the Java enum lists.
type NewsType int

const (
	NT_ARRIVAL_COMPANY NewsType = 0 // First vehicle arrived for company.
	NT_ARRIVAL_OTHER NewsType = 1 // First vehicle arrived for competitor.
	NT_ACCIDENT NewsType = 2 // An accident or disaster has occurred.
	NT_COMPANY_INFO NewsType = 3 // Company info: new companies, bankruptcies.
	NT_INDUSTRY_OPEN NewsType = 4 // Opening of industries.
	NT_INDUSTRY_CLOSE NewsType = 5 // Closing of industries.
	NT_ECONOMY NewsType = 6 // Economic changes: recession, industry up/down.
	NT_INDUSTRY_COMPANY NewsType = 7 // Production changes of an industry serviced by the local company.
	NT_INDUSTRY_OTHER NewsType = 8 // Production changes of an industry serviced by competitors.
	NT_INDUSTRY_NOBODY NewsType = 9 // Other industry production changes.
	NT_ADVICE NewsType = 10 // Bits of news about the vehicles of the company.
	NT_NEW_VEHICLES NewsType = 11 // A new vehicle has become available.
	NT_ACCEPTANCE NewsType = 12 // A type of cargo is (no longer) accepted.
	NT_SUBSIDIES NewsType = 13 // News about subsidies: announcements, expirations, acceptance.
	NT_GENERAL NewsType = 14 // General news, from towns.
)

const NEWS_TYPE_COUNT = 15

// IsValid is NewsType.isValid(). As of OpenTTD 1.3.0 the gamescript only
// accepts general, economy and subsidies, so everything else is replaced by
// NT_GENERAL before it goes on the wire.
func (t NewsType) IsValid() bool {
	if t == NT_GENERAL || t == NT_ECONOMY || t == NT_SUBSIDIES {
		return true
	}
	return false
}

// NewsTypeName is the enum's name, for the logs.
func NewsTypeName(newsType NewsType) string {
	if newsType < 0 || newsType >= NEWS_TYPE_COUNT {
		return "NEWS_TYPE_UNKNOWN"
	}
	return newsTypeNames[int(newsType)]
}

var newsTypeNames = []string{
	"NT_ARRIVAL_COMPANY",
	"NT_ARRIVAL_OTHER",
	"NT_ACCIDENT",
	"NT_COMPANY_INFO",
	"NT_INDUSTRY_OPEN",
	"NT_INDUSTRY_CLOSE",
	"NT_ECONOMY",
	"NT_INDUSTRY_COMPANY",
	"NT_INDUSTRY_OTHER",
	"NT_INDUSTRY_NOBODY",
	"NT_ADVICE",
	"NT_NEW_VEHICLES",
	"NT_ACCEPTANCE",
	"NT_SUBSIDIES",
	"NT_GENERAL",
}

// NewsTypeOf is NewsType.valueOf(int).
func NewsTypeOf(order int) NewsType {
	if order < 0 || order >= NEWS_TYPE_COUNT {
		return NEWS_TYPE_UNKNOWN
	}
	return NewsType(order)
}

const NEWS_TYPE_UNKNOWN NewsType = -1

type GSNewsPaper struct {
	company int
	newsType NewsType
	message string
	id string
	response *GSResponse
}

// NewGSNewsPaper is GSNewsPaper(NewsType, String): a newspaper for every
// company.
func NewGSNewsPaper(newsType NewsType, message string) *GSNewsPaper {
	return &GSNewsPaper{company: -1, newsType: newsType, message: message}
}

// NewGSNewsPaperForCompany is GSNewsPaper(short, NewsType, String): a newspaper
// for one company.
func NewGSNewsPaperForCompany(company int, newsType NewsType, message string) *GSNewsPaper {
	return &GSNewsPaper{company: company, newsType: newsType, message: message}
}

// GetCompany is the companyId field.
func (r *GSNewsPaper) GetCompany() int { return r.company }

// GetId, SetId, HasResponse and SetResponse are the GSRequest base class.
func (r *GSNewsPaper) GetId() string { return r.id }

func (r *GSNewsPaper) SetId(id string) { r.id = id }

func (r *GSNewsPaper) HasResponse() bool { return r.response != nil }

func (r *GSNewsPaper) SetResponse(response *GSResponse) { r.response = response }

// ToJson is toJson. Java replaces an unusable news type by NT_GENERAL by
// assigning to the field, so the replacement is still visible afterwards; the
// Go port keeps that behaviour.
func (r *GSNewsPaper) ToJson() string {
	if !r.newsType.IsValid() {
		r.newsType = NT_GENERAL
	}
	var arguments = []string{
		JsonNumber(int(r.newsType)),
		JsonString(r.message),
		JsonNumber(r.company),
	}
	var fields = []JsonField{
		JsonField{Key: constant.GS_ID, Value: JsonString(r.id)},
		JsonField{Key: constant.GS_CMD, Value: JsonNumber(int(constant.GS_COMMAND_CREATE_NEWS))},
		JsonField{Key: constant.GS_ARGS, Value: JsonList(arguments)},
	}
	return JsonObject(fields)
}

// ToString is toString: the wire payload.
func (r *GSNewsPaper) ToString() string { return r.ToJson() }

// GetResult is fromJson, Boolean.valueOf of the reply: true only when the reply
// reads "true", in any case.
func (r *GSNewsPaper) GetResult() (bool, *GSException) {
	if r.response == nil {
		return false, NewException("No gamescript reply for request " + r.id)
	}
	if r.response.HasException {
		return false, NewException(r.response.Exception)
	}
	return strings.ToLower(r.response.Data) == "true", nil
}
