require("version.nut");

class FMainClass extends GSInfo {
	function GetAuthor()		{ return "Eric Renard"; }
	function GetName()		{ return "AdminCmd"; }
	function GetDescription() 	{ return "Server script expanding the admin port commands"; }
	function GetVersion()		{ return SELF_VERSION; }
	function GetDate()		{ return "2013-04-10"; }
	function CreateInstance()	{ return "MainClass"; }
	function GetShortName()		{ return "ADMC"; }
	function GetAPIVersion()	{ return "1.2"; }
	function GetUrl()		{ return "https://github.com/erenard"; }
	function GetSettings() {
		AddSetting({
			name = "log_level", 
			description = "Debug: Log level (higher = print more)", 
			easy_value = 3, 
			medium_value = 3, 
			hard_value = 3, 
			custom_value = 3, 
			flags = CONFIG_INGAME, 
			min_value = 1, 
			max_value = 3
		});
	}
}

RegisterGS(FMainClass());
