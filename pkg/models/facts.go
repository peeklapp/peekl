package models

type PackageFact struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type NetworkInterfaceAddressInfo struct {
	Family           string `json:"family"`
	Address          string `json:"local"`
	PrefixLen        string `json:"prefix_len"`
	BroadcastAddress string `json:"broadcast"`
}

type NetworkInterface struct {
	Name        string   `json:"ifname"`
	Type        string   `json:"type"`
	Flags       []string `json:"flags"`
	Mtu         int      `json:"mtu"`
	MacAddress  string   `json:"address"`
	AddressInfo []NetworkInterfaceAddressInfo
}

type LsbData struct {
	DistributorId string `json:"distributor_id"`
	Description   string `json:"description"`
	Release       string `json:"release"`
	Codename      string `json:"codename"`
}

type DmiBiosData struct {
	Date    string `json:"date" sysfs:"bios_date" mapstructure:"bios_date"`
	Release string `json:"release" sysfs:"bios_release" mapstructure:"bios_release"`
	Vendor  string `json:"vendor" sysfs:"bios_vendor" mapstructure:"bios_vendor"`
	Version string `json:"version" sysfs:"bios_version" mapstructure:"bios_version"`
}

type DmiBoardData struct {
	AssetTag string `json:"asset_tag" sysfs:"board_asset_tag" mapstructure:"chassis_asset_tag"`
	Name     string `json:"name" sysfs:"board_name" mapstructure:"chassis_name"`
	Serial   string `json:"serial" sysfs:"board_serial" mapstructure:"chassis_serial"`
	Vendor   string `json:"vendor" sysfs:"board_vendor" mapstructure:"chassis_vendor"`
	Version  string `json:"version" sysfs:"board_version" mapstructure:"board_version"`
}

type DmiChassisData struct {
	AssetTag string `json:"asset_tag" sysfs:"chassis_asset_tag" mapstructure:"chassis_asset_tag"`
	Serial   string `json:"serial" sysfs:"chassis_serial" mapstructure:"chassis_serial"`
	Type     string `json:"type" sysfs:"chassis_type" mapstructure:"chassis_type"`
	Vendor   string `json:"vendor" sysfs:"chassis_vendor" mapstructure:"chassis_vendor"`
	Version  string `json:"version" sysfs:"chassis_version" mapstructure:"chassis_version"`
}

type DmiProductData struct {
	Family  string `json:"family" sysfs:"product_family" mapstructure:"product_family"`
	Name    string `json:"name" sysfs:"product_name" mapstructure:"product_name"`
	Serial  string `json:"serial" sysfs:"product_serial" mapstructure:"product_serial"`
	Sku     string `json:"sku" sysfs:"product_sku" mapstructure:"product_sku"`
	Uuid    string `json:"uuid" sysfs:"product_uuid" mapstructure:"product_uuid"`
	Version string `json:"version" sysfs:"product_version" mapstructure:"product_version"`
}

type DmiData struct {
	Bios    DmiBiosData    `json:"bios" mapstructure:"bios"`
	Board   DmiBoardData   `json:"board" mapstructure:"board"`
	Chassis DmiChassisData `json:"chassis" mapstructure:"chassis"`
	Product DmiProductData `json:"product" mapstructure:"product"`
}

type DiskPartition struct {
	Name       string   `json:"name"`
	Size       string   `json:"size"`
	Type       string   `json:"type"`
	Mountpoint []string `json:"mountpoints"`
}

type Disk struct {
	Name        string          `json:"name"`
	Size        string          `json:"size"`
	Type        string          `json:"type"`
	Mountpoints []string        `json:"mountpoints"`
	Partitions  []DiskPartition `json:"children,omitempty"`
}

type Facts struct {
	Hostname          string             `json:"hostname"`
	Packages          []Package          `json:"packages"`
	NetworkInterfaces []NetworkInterface `json:"network_interfaces"`
	Lsb               LsbData            `json:"lsb"`
	Dmi               DmiData            `json:"dmi"`
	Disks             []Disk             `json:"disks"`
}
