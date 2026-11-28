// Client mirrors com.openttd.network.admin.Client: one game client as the server
// describes it in ADMIN_PACKET_SERVER_CLIENT_INFO. Java's final long id becomes
// an int field set by the constructor, and there is no setId, as in Java.

package admin

import "strconv"

type Client struct {
	id       int
	ip       string
	name     string
	joinDate int64
	company  int
}

// NewClient is the constructor: it takes the id, everything else is filled in
// by NetworkAdmin.handlePacket.
func NewClient(id int) *Client {
	return &Client{id: id}
}

// The getters, in Java's order.
func (c *Client) GetIp() string      { return c.ip }
func (c *Client) GetName() string    { return c.name }
func (c *Client) GetJoinDate() int64 { return c.joinDate }
func (c *Client) GetCompany() int    { return c.company }
func (c *Client) GetId() int         { return c.id }

// The setters.
func (c *Client) SetIp(ip string)            { c.ip = ip }
func (c *Client) SetName(name string)        { c.name = name }
func (c *Client) SetJoinDate(joinDate int64) { c.joinDate = joinDate }
func (c *Client) SetCompany(company int)     { c.company = company }

// ToString is toString: id + ":" + name + ", " + ip + ", " + companyId.
func (c *Client) ToString() string {
	return strconv.Itoa(c.id) + ":" + c.name + ", " + c.ip + ", " + strconv.Itoa(c.company)
}

// Copy is clone.
func (c *Client) Copy() *Client {
	return &Client{
		id:       c.id,
		ip:       c.ip,
		name:     c.name,
		joinDate: c.joinDate,
		company:  c.company,
	}
}
