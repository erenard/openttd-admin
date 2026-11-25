// Company mirrors com.openttd.network.admin.Company: everything the COMPANY_INFO,
// COMPANY_UPDATE, COMPANY_ECONOMY and COMPANY_STATS packets describe. Java's
// short and char fields become int, its long becomes int64, and its three
// BigInteger fields become uint64, because that is what Packet.ReadUint64
// produces and OpenTTD's economy values are unsigned. Java leaves them null
// until an economy packet arrives; here they read as 0, so a Go caller never
// has to repeat the null checks CompanyEconomyObjective needs.
//
// Java's short[4] shareOwners and char[5] vehicules and stations become []int,
// created with the same default lengths, and Copy deep-copies them as Java's
// clone() does.

package admin

import "strconv"

type Company struct {
	id int

	name string
	president string
	color int
	usePassword bool
	inauguratedYear int64
	ai bool

	bankruptcy int
	shareOwners []int

	money int64
	loan uint64
	income int64

	lastValue uint64
	previousValue uint64

	lastPerformance int
	previousPerformance int

	deliveredCargo int
	lastDeliveredCargo int
	previousDeliveredCargo int

	vehicules []int
	stations []int
}

// NewCompany is the constructor: it takes the id, and the three arrays get their
// default lengths, as in Java's field initializers.
func NewCompany(id int) *Company {
	return &Company{
		id: id,
		shareOwners: newIntList(4),
		vehicules: newIntList(5),
		stations: newIntList(5),
	}
}

// The getters, in Java's order.
func (c *Company) GetId() int { return c.id }
func (c *Company) GetName() string { return c.name }
func (c *Company) GetPresident() string { return c.president }
func (c *Company) GetColor() int { return c.color }
func (c *Company) IsUsePassword() bool { return c.usePassword }
func (c *Company) GetInauguratedYear() int64 { return c.inauguratedYear }
func (c *Company) IsAi() bool { return c.ai }
func (c *Company) GetBankruptcy() int { return c.bankruptcy }
func (c *Company) GetShareOwners() []int { return c.shareOwners }
func (c *Company) GetMoney() int64 { return c.money }
func (c *Company) GetLoan() uint64 { return c.loan }
func (c *Company) GetIncome() int64 { return c.income }
func (c *Company) GetLastValue() uint64 { return c.lastValue }
func (c *Company) GetPreviousValue() uint64 { return c.previousValue }
func (c *Company) GetLastPerformance() int { return c.lastPerformance }
func (c *Company) GetPreviousPerformance() int { return c.previousPerformance }
func (c *Company) GetDeliveredCargo() int { return c.deliveredCargo }
func (c *Company) GetLastDeliveredCargo() int { return c.lastDeliveredCargo }
func (c *Company) GetPreviousDeliveredCargo() int { return c.previousDeliveredCargo }
func (c *Company) GetVehicules() []int { return c.vehicules }
func (c *Company) GetStations() []int { return c.stations }

// The setters. Java has no setId: the id is final, set by the constructor.
func (c *Company) SetName(name string) { c.name = name }
func (c *Company) SetPresident(president string) { c.president = president }
func (c *Company) SetColor(color int) { c.color = color }
func (c *Company) SetUsePassword(usePassword bool) { c.usePassword = usePassword }
func (c *Company) SetInauguratedYear(inauguratedYear int64) { c.inauguratedYear = inauguratedYear }
func (c *Company) SetAi(ai bool) { c.ai = ai }
func (c *Company) SetBankruptcy(bankruptcy int) { c.bankruptcy = bankruptcy }
func (c *Company) SetShareOwners(shareOwners []int) { c.shareOwners = shareOwners }
func (c *Company) SetMoney(money int64) { c.money = money }
func (c *Company) SetLoan(loan uint64) { c.loan = loan }
func (c *Company) SetIncome(income int64) { c.income = income }
func (c *Company) SetLastValue(lastValue uint64) { c.lastValue = lastValue }
func (c *Company) SetPreviousValue(previousValue uint64) { c.previousValue = previousValue }
func (c *Company) SetLastPerformance(lastPerformance int) { c.lastPerformance = lastPerformance }
func (c *Company) SetPreviousPerformance(previousPerformance int) { c.previousPerformance = previousPerformance }
func (c *Company) SetDeliveredCargo(deliveredCargo int) { c.deliveredCargo = deliveredCargo }
func (c *Company) SetLastDeliveredCargo(lastDeliveredCargo int) { c.lastDeliveredCargo = lastDeliveredCargo }
func (c *Company) SetPreviousDeliveredCargo(previousDeliveredCargo int) { c.previousDeliveredCargo = previousDeliveredCargo }
func (c *Company) SetVehicules(vehicules []int) { c.vehicules = vehicules }
func (c *Company) SetStations(stations []int) { c.stations = stations }

// ToString is toString: id + ":" + name + ", " + inauguratedYear + ", password:"
// + usePassword, then money, income, lastValue and lastPerformance tab separated.
func (c *Company) ToString() string {
	return strconv.Itoa(c.id) + ":" + c.name + ", " + strconv.FormatInt(c.inauguratedYear, 10) +
		", password:" + boolText(c.usePassword) +
		"\t" + strconv.FormatInt(c.money, 10) +
		"\t" + strconv.FormatInt(c.income, 10) +
		"\t" + strconv.FormatUint(c.lastValue, 10) +
		"\t" + strconv.Itoa(c.lastPerformance)
}

// Copy is clone: a new Company with the same id, every field carried over, and
// fresh copies of the three arrays.
func (c *Company) Copy() *Company {
	var clone = &Company{
		id: c.id,
		name: c.name,
		president: c.president,
		color: c.color,
		usePassword: c.usePassword,
		inauguratedYear: c.inauguratedYear,
		ai: c.ai,
		bankruptcy: c.bankruptcy,
		shareOwners: copyIntList(c.shareOwners),
		money: c.money,
		loan: c.loan,
		income: c.income,
		lastValue: c.lastValue,
		previousValue: c.previousValue,
		lastPerformance: c.lastPerformance,
		previousPerformance: c.previousPerformance,
		deliveredCargo: c.deliveredCargo,
		lastDeliveredCargo: c.lastDeliveredCargo,
		previousDeliveredCargo: c.previousDeliveredCargo,
		vehicules: copyIntList(c.vehicules),
		stations: copyIntList(c.stations),
	}
	return clone
}

// newIntList is Java's new short[4] or new char[5]: zeros of the right length.
func newIntList(length int) []int {
	var list []int
	for i := 0; i < length; i++ {
		list = append(list, 0)
	}
	return list
}

// copyIntList is array.clone().
func copyIntList(list []int) []int {
	var copy []int
	for i := 0; i < len(list); i++ {
		copy = append(copy, list[i])
	}
	return copy
}

