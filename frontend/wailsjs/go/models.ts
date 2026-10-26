export namespace runtime {
	
	export class Result {
	    Stdout: string;
	    Stderr: string;
	    ExitCode: number;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Stdout = source["Stdout"];
	        this.Stderr = source["Stderr"];
	        this.ExitCode = source["ExitCode"];
	    }
	}

}

