export namespace port {
	
	export class ConnectionInfo {
	    local_ip: string;
	    public_ip: string;
	    port: number;
	    upnp_ok: boolean;
	    local_url: string;
	    public_url: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.local_ip = source["local_ip"];
	        this.public_ip = source["public_ip"];
	        this.port = source["port"];
	        this.upnp_ok = source["upnp_ok"];
	        this.local_url = source["local_url"];
	        this.public_url = source["public_url"];
	    }
	}

}

