export namespace models {
	
	export class PatientTreatment {
	    id: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    patientId: string;
	    treatmentId: string;
	    invoiceId: string;
	    treatmentDate: string;
	    toothNumber: string;
	    notes: string;
	    performedBy: string;
	    patient?: Patient;
	    treatment?: Treatment;
	
	    static createFrom(source: any = {}) {
	        return new PatientTreatment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.patientId = source["patientId"];
	        this.treatmentId = source["treatmentId"];
	        this.invoiceId = source["invoiceId"];
	        this.treatmentDate = source["treatmentDate"];
	        this.toothNumber = source["toothNumber"];
	        this.notes = source["notes"];
	        this.performedBy = source["performedBy"];
	        this.patient = this.convertValues(source["patient"], Patient);
	        this.treatment = this.convertValues(source["treatment"], Treatment);
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
	export class Payment {
	    id: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    invoiceId: string;
	    amount: number;
	    method: string;
	    paymentDate: string;
	    reference: string;
	    notes: string;
	    receivedBy: string;
	    invoice?: Invoice;
	
	    static createFrom(source: any = {}) {
	        return new Payment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.invoiceId = source["invoiceId"];
	        this.amount = source["amount"];
	        this.method = source["method"];
	        this.paymentDate = source["paymentDate"];
	        this.reference = source["reference"];
	        this.notes = source["notes"];
	        this.receivedBy = source["receivedBy"];
	        this.invoice = this.convertValues(source["invoice"], Invoice);
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
	export class Treatment {
	    id: string;
	    name: string;
	    code: string;
	    defaultPrice: number;
	    category: string;
	    description: string;
	    isActive: boolean;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Treatment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.code = source["code"];
	        this.defaultPrice = source["defaultPrice"];
	        this.category = source["category"];
	        this.description = source["description"];
	        this.isActive = source["isActive"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class InvoiceItem {
	    id: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    invoiceId: string;
	    treatmentId?: string;
	    description: string;
	    quantity: number;
	    unitPrice: number;
	    amount: number;
	    toothNumber: string;
	    treatment?: Treatment;
	
	    static createFrom(source: any = {}) {
	        return new InvoiceItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.invoiceId = source["invoiceId"];
	        this.treatmentId = source["treatmentId"];
	        this.description = source["description"];
	        this.quantity = source["quantity"];
	        this.unitPrice = source["unitPrice"];
	        this.amount = source["amount"];
	        this.toothNumber = source["toothNumber"];
	        this.treatment = this.convertValues(source["treatment"], Treatment);
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
	export class Invoice {
	    id: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    invoiceNumber: string;
	    patientId: string;
	    invoiceDate: string;
	    subTotal: number;
	    discountAmount: number;
	    discountPercent: number;
	    taxableAmount: number;
	    cgstAmount: number;
	    sgstAmount: number;
	    totalAmount: number;
	    paidAmount: number;
	    balanceAmount: number;
	    status: string;
	    notes: string;
	    voidReason: string;
	    createdBy: string;
	    patient?: Patient;
	    items?: InvoiceItem[];
	    payments?: Payment[];
	
	    static createFrom(source: any = {}) {
	        return new Invoice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.invoiceNumber = source["invoiceNumber"];
	        this.patientId = source["patientId"];
	        this.invoiceDate = source["invoiceDate"];
	        this.subTotal = source["subTotal"];
	        this.discountAmount = source["discountAmount"];
	        this.discountPercent = source["discountPercent"];
	        this.taxableAmount = source["taxableAmount"];
	        this.cgstAmount = source["cgstAmount"];
	        this.sgstAmount = source["sgstAmount"];
	        this.totalAmount = source["totalAmount"];
	        this.paidAmount = source["paidAmount"];
	        this.balanceAmount = source["balanceAmount"];
	        this.status = source["status"];
	        this.notes = source["notes"];
	        this.voidReason = source["voidReason"];
	        this.createdBy = source["createdBy"];
	        this.patient = this.convertValues(source["patient"], Patient);
	        this.items = this.convertValues(source["items"], InvoiceItem);
	        this.payments = this.convertValues(source["payments"], Payment);
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
	export class Patient {
	    id: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    name: string;
	    phone: string;
	    email: string;
	    gender: string;
	    age: number;
	    dateOfBirth: string;
	    address: string;
	    city: string;
	    bloodGroup: string;
	    medicalHistory: string;
	    allergies: string;
	    notes: string;
	    createdBy: string;
	    appointments?: Appointment[];
	    invoices?: Invoice[];
	    patientTreatments?: PatientTreatment[];
	
	    static createFrom(source: any = {}) {
	        return new Patient(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.name = source["name"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.gender = source["gender"];
	        this.age = source["age"];
	        this.dateOfBirth = source["dateOfBirth"];
	        this.address = source["address"];
	        this.city = source["city"];
	        this.bloodGroup = source["bloodGroup"];
	        this.medicalHistory = source["medicalHistory"];
	        this.allergies = source["allergies"];
	        this.notes = source["notes"];
	        this.createdBy = source["createdBy"];
	        this.appointments = this.convertValues(source["appointments"], Appointment);
	        this.invoices = this.convertValues(source["invoices"], Invoice);
	        this.patientTreatments = this.convertValues(source["patientTreatments"], PatientTreatment);
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
	export class Appointment {
	    id: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    patientId: string;
	    appointmentDate: string;
	    startTime: string;
	    endTime: string;
	    duration: number;
	    status: string;
	    purpose: string;
	    notes: string;
	    cancelReason: string;
	    createdBy: string;
	    patient?: Patient;
	
	    static createFrom(source: any = {}) {
	        return new Appointment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.patientId = source["patientId"];
	        this.appointmentDate = source["appointmentDate"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	        this.duration = source["duration"];
	        this.status = source["status"];
	        this.purpose = source["purpose"];
	        this.notes = source["notes"];
	        this.cancelReason = source["cancelReason"];
	        this.createdBy = source["createdBy"];
	        this.patient = this.convertValues(source["patient"], Patient);
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
	export class ClinicSettings {
	    id: string;
	    clinicName: string;
	    doctorName: string;
	    doctorQualification: string;
	    address: string;
	    city: string;
	    state: string;
	    pincode: string;
	    phone: string;
	    email: string;
	    gstin: string;
	    gstEnabled: boolean;
	    gstRate: number;
	    invoicePrefix: string;
	    logoPath: string;
	    logoBase64: string;
	    setupComplete: boolean;
	    autoBackup: boolean;
	    backupPath: string;
	    cloudBackupEnabled: boolean;
	    cloudBackupPath: string;
	    bankAccount: string;
	    accountName: string;
	    bankName: string;
	    ifscCode: string;
	    upiId: string;
	    whatsAppEnabled: boolean;
	    whatsAppWelcomeTemplate: string;
	    whatsAppInvoiceTemplate: string;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new ClinicSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.clinicName = source["clinicName"];
	        this.doctorName = source["doctorName"];
	        this.doctorQualification = source["doctorQualification"];
	        this.address = source["address"];
	        this.city = source["city"];
	        this.state = source["state"];
	        this.pincode = source["pincode"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.gstin = source["gstin"];
	        this.gstEnabled = source["gstEnabled"];
	        this.gstRate = source["gstRate"];
	        this.invoicePrefix = source["invoicePrefix"];
	        this.logoPath = source["logoPath"];
	        this.logoBase64 = source["logoBase64"];
	        this.setupComplete = source["setupComplete"];
	        this.autoBackup = source["autoBackup"];
	        this.backupPath = source["backupPath"];
	        this.cloudBackupEnabled = source["cloudBackupEnabled"];
	        this.cloudBackupPath = source["cloudBackupPath"];
	        this.bankAccount = source["bankAccount"];
	        this.accountName = source["accountName"];
	        this.bankName = source["bankName"];
	        this.ifscCode = source["ifscCode"];
	        this.upiId = source["upiId"];
	        this.whatsAppEnabled = source["whatsAppEnabled"];
	        this.whatsAppWelcomeTemplate = source["whatsAppWelcomeTemplate"];
	        this.whatsAppInvoiceTemplate = source["whatsAppInvoiceTemplate"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	
	
	
	
	

}

export namespace service {
	
	export class UserInfo {
	    id: string;
	    username: string;
	    fullName: string;
	    role: string;
	
	    static createFrom(source: any = {}) {
	        return new UserInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.fullName = source["fullName"];
	        this.role = source["role"];
	    }
	}
	export class AuthResponse {
	    user: UserInfo;
	    loggedIn: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AuthResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user = this.convertValues(source["user"], UserInfo);
	        this.loggedIn = source["loggedIn"];
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
	export class BackupInfo {
	    fileName: string;
	    filePath: string;
	    size: number;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fileName = source["fileName"];
	        this.filePath = source["filePath"];
	        this.size = source["size"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class CloudDriveInfo {
	    provider: string;
	    path: string;
	    available: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CloudDriveInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.path = source["path"];
	        this.available = source["available"];
	    }
	}
	export class CreateAppointmentInput {
	    patientId: string;
	    date: string;
	    startTime: string;
	    endTime: string;
	    duration: number;
	    purpose: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateAppointmentInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.patientId = source["patientId"];
	        this.date = source["date"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	        this.duration = source["duration"];
	        this.purpose = source["purpose"];
	        this.notes = source["notes"];
	    }
	}
	export class InvoiceItemInput {
	    treatmentId: string;
	    description: string;
	    quantity: number;
	    unitPrice: number;
	    toothNumber: string;
	
	    static createFrom(source: any = {}) {
	        return new InvoiceItemInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.treatmentId = source["treatmentId"];
	        this.description = source["description"];
	        this.quantity = source["quantity"];
	        this.unitPrice = source["unitPrice"];
	        this.toothNumber = source["toothNumber"];
	    }
	}
	export class CreateInvoiceInput {
	    patientId: string;
	    items: InvoiceItemInput[];
	    discountPercent: number;
	    discountAmount: number;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateInvoiceInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.patientId = source["patientId"];
	        this.items = this.convertValues(source["items"], InvoiceItemInput);
	        this.discountPercent = source["discountPercent"];
	        this.discountAmount = source["discountAmount"];
	        this.notes = source["notes"];
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
	export class CreatePatientInput {
	    name: string;
	    phone: string;
	    email: string;
	    gender: string;
	    age: number;
	    dateOfBirth: string;
	    address: string;
	    city: string;
	    bloodGroup: string;
	    medicalHistory: string;
	    allergies: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new CreatePatientInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.gender = source["gender"];
	        this.age = source["age"];
	        this.dateOfBirth = source["dateOfBirth"];
	        this.address = source["address"];
	        this.city = source["city"];
	        this.bloodGroup = source["bloodGroup"];
	        this.medicalHistory = source["medicalHistory"];
	        this.allergies = source["allergies"];
	        this.notes = source["notes"];
	    }
	}
	export class PaymentSummary {
	    invoiceNumber: string;
	    patientName: string;
	    amount: number;
	    method: string;
	
	    static createFrom(source: any = {}) {
	        return new PaymentSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.invoiceNumber = source["invoiceNumber"];
	        this.patientName = source["patientName"];
	        this.amount = source["amount"];
	        this.method = source["method"];
	    }
	}
	export class DailyReport {
	    date: string;
	    totalCollection: number;
	    payments: PaymentSummary[];
	
	    static createFrom(source: any = {}) {
	        return new DailyReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.totalCollection = source["totalCollection"];
	        this.payments = this.convertValues(source["payments"], PaymentSummary);
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
	export class DashboardStats {
	    todayAppointments: number;
	    totalPatients: number;
	    todayRevenue: number;
	    monthRevenue: number;
	    totalOutstanding: number;
	    patientsThisMonth: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.todayAppointments = source["todayAppointments"];
	        this.totalPatients = source["totalPatients"];
	        this.todayRevenue = source["todayRevenue"];
	        this.monthRevenue = source["monthRevenue"];
	        this.totalOutstanding = source["totalOutstanding"];
	        this.patientsThisMonth = source["patientsThisMonth"];
	    }
	}
	
	export class InvoiceListResponse {
	    invoices: models.Invoice[];
	    total: number;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new InvoiceListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.invoices = this.convertValues(source["invoices"], models.Invoice);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
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
	export class MonthlyReport {
	    year: number;
	    month: number;
	    totalRevenue: number;
	    totalInvoiced: number;
	    totalOutstanding: number;
	
	    static createFrom(source: any = {}) {
	        return new MonthlyReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.year = source["year"];
	        this.month = source["month"];
	        this.totalRevenue = source["totalRevenue"];
	        this.totalInvoiced = source["totalInvoiced"];
	        this.totalOutstanding = source["totalOutstanding"];
	    }
	}
	export class PatientListResponse {
	    patients: models.Patient[];
	    total: number;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new PatientListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.patients = this.convertValues(source["patients"], models.Patient);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
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
	
	export class RecordPaymentInput {
	    invoiceId: string;
	    amount: number;
	    method: string;
	    paymentDate: string;
	    reference: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new RecordPaymentInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.invoiceId = source["invoiceId"];
	        this.amount = source["amount"];
	        this.method = source["method"];
	        this.paymentDate = source["paymentDate"];
	        this.reference = source["reference"];
	        this.notes = source["notes"];
	    }
	}
	export class SetupInput {
	    clinicName: string;
	    doctorName: string;
	    doctorQualification: string;
	    address: string;
	    city: string;
	    state: string;
	    pincode: string;
	    phone: string;
	    email: string;
	    gstin: string;
	    gstEnabled: boolean;
	    invoicePrefix: string;
	    adminUsername: string;
	    adminPassword: string;
	    adminFullName: string;
	
	    static createFrom(source: any = {}) {
	        return new SetupInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.clinicName = source["clinicName"];
	        this.doctorName = source["doctorName"];
	        this.doctorQualification = source["doctorQualification"];
	        this.address = source["address"];
	        this.city = source["city"];
	        this.state = source["state"];
	        this.pincode = source["pincode"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.gstin = source["gstin"];
	        this.gstEnabled = source["gstEnabled"];
	        this.invoicePrefix = source["invoicePrefix"];
	        this.adminUsername = source["adminUsername"];
	        this.adminPassword = source["adminPassword"];
	        this.adminFullName = source["adminFullName"];
	    }
	}
	export class UpdateInfo {
	    available: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    downloadURL: string;
	    releaseNotes: string;
	    publishedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.downloadURL = source["downloadURL"];
	        this.releaseNotes = source["releaseNotes"];
	        this.publishedAt = source["publishedAt"];
	    }
	}
	
	export class WhatsAppMessageResult {
	    phone: string;
	    message: string;
	    whatsAppUrl: string;
	    webUrl: string;
	    isDesktopPresent: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WhatsAppMessageResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.phone = source["phone"];
	        this.message = source["message"];
	        this.whatsAppUrl = source["whatsAppUrl"];
	        this.webUrl = source["webUrl"];
	        this.isDesktopPresent = source["isDesktopPresent"];
	    }
	}
	export class WhatsAppTemplates {
	    welcomeTemplate: string;
	    invoiceTemplate: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WhatsAppTemplates(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.welcomeTemplate = source["welcomeTemplate"];
	        this.invoiceTemplate = source["invoiceTemplate"];
	        this.enabled = source["enabled"];
	    }
	}

}

