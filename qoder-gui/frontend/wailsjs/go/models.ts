export namespace main {
	
	export class BackupInfo {
	    name: string;
	    path: string;
	    size: number;
	    modTime: number;
	    timeStr: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.modTime = source["modTime"];
	        this.timeStr = source["timeStr"];
	    }
	}
	export class BackupResult {
	    success: boolean;
	    message: string;
	    backupPath?: string;
	    count: number;
	    successCount: number;
	
	    static createFrom(source: any = {}) {
	        return new BackupResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.backupPath = source["backupPath"];
	        this.count = source["count"];
	        this.successCount = source["successCount"];
	    }
	}
	export class ChatInfo {
	    id: string;
	    title: string;
	    timestamp: number;
	    sessionId: string;
	    timeStr: string;
	
	    static createFrom(source: any = {}) {
	        return new ChatInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.timestamp = source["timestamp"];
	        this.sessionId = source["sessionId"];
	        this.timeStr = source["timeStr"];
	    }
	}
	export class RestoreResult {
	    success: boolean;
	    message: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new RestoreResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class WorkspaceInfo {
	    id: string;
	    name: string;
	    path: string;
	    chatCount: number;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.chatCount = source["chatCount"];
	    }
	}

}

