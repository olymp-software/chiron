export namespace apperr {
	
	export class Problem {
	    code: string;
	    path?: string;
	    params?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Problem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.path = source["path"];
	        this.params = source["params"];
	    }
	}
	export class Error {
	    code: string;
	    params?: Record<string, string>;
	    problems?: Problem[];
	
	    static createFrom(source: any = {}) {
	        return new Error(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.params = source["params"];
	        this.problems = this.convertValues(source["problems"], Problem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace library {
	
	export class ActivityControl {
	    label: string;
	    group: string;
	
	    static createFrom(source: any = {}) {
	        return new ActivityControl(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.group = source["group"];
	    }
	}
	export class WorkshopActivity {
	    solution?: string;
	    hasCheck: boolean;
	    controls: ActivityControl[];
	
	    static createFrom(source: any = {}) {
	        return new WorkshopActivity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.solution = source["solution"];
	        this.hasCheck = source["hasCheck"];
	        this.controls = this.convertValues(source["controls"], ActivityControl);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TheoryBlock {
	    kind: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new TheoryBlock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.content = source["content"];
	    }
	}
	export class Activity {
	    slug: string;
	    title: string;
	    dir: string;
	    type: string;
	    theoryBlocks: TheoryBlock[];
	    workshop?: WorkshopActivity;
	    last: boolean;
	    courseCompleted: boolean;
	    error?: apperr.Error;
	
	    static createFrom(source: any = {}) {
	        return new Activity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.slug = source["slug"];
	        this.title = source["title"];
	        this.dir = source["dir"];
	        this.type = source["type"];
	        this.theoryBlocks = this.convertValues(source["theoryBlocks"], TheoryBlock);
	        this.workshop = this.convertValues(source["workshop"], WorkshopActivity);
	        this.last = source["last"];
	        this.courseCompleted = source["courseCompleted"];
	        this.error = this.convertValues(source["error"], apperr.Error);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Course {
	    origin: string;
	    source: string;
	    key: string;
	    title: string;
	    type?: string;
	    checkHash?: string;
	    enrolled: boolean;
	    completed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Course(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.origin = source["origin"];
	        this.source = source["source"];
	        this.key = source["key"];
	        this.title = source["title"];
	        this.type = source["type"];
	        this.checkHash = source["checkHash"];
	        this.enrolled = source["enrolled"];
	        this.completed = source["completed"];
	    }
	}
	export class Entry {
	    origin: string;
	    source: string;
	    key: string;
	    title: string;
	    type?: string;
	    checkHash?: string;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.origin = source["origin"];
	        this.source = source["source"];
	        this.key = source["key"];
	        this.title = source["title"];
	        this.type = source["type"];
	        this.checkHash = source["checkHash"];
	    }
	}
	export class Overview {
	    origin: string;
	    source: string;
	    key: string;
	    title: string;
	    type?: string;
	    checkHash?: string;
	    enrolled: boolean;
	    completed: boolean;
	    text: manifest.OverviewText;
	    check: string;
	    controls?: string[];
	    error?: apperr.Error;
	
	    static createFrom(source: any = {}) {
	        return new Overview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.origin = source["origin"];
	        this.source = source["source"];
	        this.key = source["key"];
	        this.title = source["title"];
	        this.type = source["type"];
	        this.checkHash = source["checkHash"];
	        this.enrolled = source["enrolled"];
	        this.completed = source["completed"];
	        this.text = this.convertValues(source["text"], manifest.OverviewText);
	        this.check = source["check"];
	        this.controls = source["controls"];
	        this.error = this.convertValues(source["error"], apperr.Error);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class TreeActivity {
	    slug: string;
	    title: string;
	    completed: boolean;
	    locked: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TreeActivity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.slug = source["slug"];
	        this.title = source["title"];
	        this.completed = source["completed"];
	        this.locked = source["locked"];
	    }
	}
	export class TreeSection {
	    slug: string;
	    title: string;
	    activities: TreeActivity[];
	
	    static createFrom(source: any = {}) {
	        return new TreeSection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.slug = source["slug"];
	        this.title = source["title"];
	        this.activities = this.convertValues(source["activities"], TreeActivity);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Tree {
	    title: string;
	    type: string;
	    workshopFolder?: string;
	    current: string;
	    completed: boolean;
	    sections: TreeSection[];
	
	    static createFrom(source: any = {}) {
	        return new Tree(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.type = source["type"];
	        this.workshopFolder = source["workshopFolder"];
	        this.current = source["current"];
	        this.completed = source["completed"];
	        this.sections = this.convertValues(source["sections"], TreeSection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	

}

export namespace main {
	
	export class ActivityResult {
	    activity: library.Activity;
	    error?: apperr.Error;
	
	    static createFrom(source: any = {}) {
	        return new ActivityResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.activity = this.convertValues(source["activity"], library.Activity);
	        this.error = this.convertValues(source["error"], apperr.Error);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CompleteResult {
	    next: string;
	    tree: library.Tree;
	    error?: apperr.Error;
	
	    static createFrom(source: any = {}) {
	        return new CompleteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.next = source["next"];
	        this.tree = this.convertValues(source["tree"], library.Tree);
	        this.error = this.convertValues(source["error"], apperr.Error);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CourseResult {
	    course: library.Entry;
	    error?: apperr.Error;
	
	    static createFrom(source: any = {}) {
	        return new CourseResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.course = this.convertValues(source["course"], library.Entry);
	        this.error = this.convertValues(source["error"], apperr.Error);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CoursesResult {
	    courses: library.Course[];
	    error?: apperr.Error;
	
	    static createFrom(source: any = {}) {
	        return new CoursesResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.courses = this.convertValues(source["courses"], library.Course);
	        this.error = this.convertValues(source["error"], apperr.Error);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EnterResult {
	    current: string;
	    error?: apperr.Error;
	
	    static createFrom(source: any = {}) {
	        return new EnterResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.current = source["current"];
	        this.error = this.convertValues(source["error"], apperr.Error);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ErrorResult {
	    error?: apperr.Error;
	
	    static createFrom(source: any = {}) {
	        return new ErrorResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.error = this.convertValues(source["error"], apperr.Error);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OverviewResult {
	    overview: library.Overview;
	    error?: apperr.Error;
	
	    static createFrom(source: any = {}) {
	        return new OverviewResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.overview = this.convertValues(source["overview"], library.Overview);
	        this.error = this.convertValues(source["error"], apperr.Error);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StartCheckResult {
	    runId: number;
	    error?: apperr.Error;
	
	    static createFrom(source: any = {}) {
	        return new StartCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.error = this.convertValues(source["error"], apperr.Error);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TreeResult {
	    tree: library.Tree;
	    error?: apperr.Error;
	
	    static createFrom(source: any = {}) {
	        return new TreeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tree = this.convertValues(source["tree"], library.Tree);
	        this.error = this.convertValues(source["error"], apperr.Error);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace manifest {
	
	export class OverviewText {
	    markdown?: string;
	    plain?: string;
	
	    static createFrom(source: any = {}) {
	        return new OverviewText(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.markdown = source["markdown"];
	        this.plain = source["plain"];
	    }
	}

}

