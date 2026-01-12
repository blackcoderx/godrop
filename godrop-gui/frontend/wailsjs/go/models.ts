export namespace backend {
	
	export class FileEntry {
	    name: string;
	    img: string;
	    isDir: boolean;
	    size: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new FileEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.img = source["img"];
	        this.isDir = source["isDir"];
	        this.size = source["size"];
	        this.type = source["type"];
	    }
	}

}

export namespace server {
	
	export class ServerResponse {
	    ip: string;
	    port: string;
	    fullUrl: string;
	    qrCode: string;
	
	    static createFrom(source: any = {}) {
	        return new ServerResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.fullUrl = source["fullUrl"];
	        this.qrCode = source["qrCode"];
	    }
	}

}

