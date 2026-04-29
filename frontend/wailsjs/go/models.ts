export namespace model {
	
	export class Asset {
	    id: number;
	    project_id: number;
	    type: string;
	    host: string;
	    port: string;
	    sources: string[];
	    status: string;
	    status_code?: number;
	    title: string;
	    server: string;
	    tech: string;
	    probed_at: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Asset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_id = source["project_id"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.sources = source["sources"];
	        this.status = source["status"];
	        this.status_code = source["status_code"];
	        this.title = source["title"];
	        this.server = source["server"];
	        this.tech = source["tech"];
	        this.probed_at = source["probed_at"];
	        this.created_at = source["created_at"];
	    }
	}
	export class ImportResult {
	    new_ip: number;
	    new_domain: number;
	    skipped: number;
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.new_ip = source["new_ip"];
	        this.new_domain = source["new_domain"];
	        this.skipped = source["skipped"];
	    }
	}
	export class Project {
	    id: number;
	    name: string;
	    created_at: string;
	    asset_count: number;
	    ip_count: number;
	    domain_count: number;
	    alive_count: number;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.created_at = source["created_at"];
	        this.asset_count = source["asset_count"];
	        this.ip_count = source["ip_count"];
	        this.domain_count = source["domain_count"];
	        this.alive_count = source["alive_count"];
	    }
	}

}

