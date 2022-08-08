enum GSCommand {
	countIndustry,
	createNews,
	addGoal,
	removeGoal,
	removeAllGoal,
	setTownCargoGoal
}

class MainClass extends GSController {
	constructor() {}
}

function MainClass::Start() {
	// Wait for the game to start
	this.Sleep(1);
	GSLog.Info("AdminCmd::Start");
	while (true) {
		this.HandleEvents();
		this.Sleep(1);
	}
}

function MainClass::Init() {
	if (this._load_data != null) {
		// Copy loaded data from this._load_data to this.*
		// or do whatever with the loaded data
	}
}

function MainClass::HandleEvents() {
	while (GSEventController.IsEventWaiting()) {
		local e = GSEventController.GetNextEvent();
		switch (e.GetEventType()) {
			case GSEvent.ET_ADMIN_PORT:
				local data = GSEventAdminPort.Convert(e).GetObject();
				if (data != null && ("i" in data) && ("c" in data) && ("a" in data)) {
					local id = data.i;
					local cmd = data.c;
					local args = data.a;
					this.ExecuteCommand(id, cmd, args);
				} else {
					GSAdmin.Send({exception = "Malformed json."});
				}
			break;
		}
	}
}

function MainClass::ExecuteCommand(id, cmd, args) {
	try {
		GSLog.Info("Command: " + id + " " + cmd + " " + args);
		local data = false;
		switch (cmd) {
			case GSCommand.countIndustry:
				data = GSIndustry.GetIndustryCount();
			break;
			case GSCommand.createNews:
				data = GSNews.Create(args[0], args[1], args[2]);
			break;
			case GSCommand.addGoal:
				data = GSGoal.New(args[0], args[1], args[2], args[3]);
			break;
			case GSCommand.removeGoal:
				if (GSGoal.IsValidGoal(goalId)) {
					data = GSGoal.Remove(args[0]);
				}
			break;
			case GSCommand.removeAllGoal:
				for (local goalId = 0; goalId < 255; goalId += 1) {
					if (GSGoal.IsValidGoal(goalId)) {
						data = data | GSGoal.Remove(goalId);
					}
				}
			break;
			case GSCommand.setTownCargoGoal:
				local town = args[0];
				local townEffect = args[1];
				if(GSTown.IsValidTown(town) && GSCargo.IsValidTownEffect(townEffect)) {
					data = GSTown.SetCargoGoal(town, townEffect, args[2]);
				}
			break;
			default:
				throw "Unknown command";
			break;
		}
		GSLog.Info("GSAdmin.Send({id = "+id+", data = "+data+"})");
		GSAdmin.Send({id = id, data = data});
	} catch (exception) {
		GSAdmin.Send({id = id, exception = exception});
	}
}
