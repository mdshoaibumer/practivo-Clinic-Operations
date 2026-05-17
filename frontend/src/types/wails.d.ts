import type { AuthResponse, PatientListResponse, InvoiceListResponse, CreatePatientInput, CreateAppointmentInput, CreateInvoiceInput, RecordPaymentInput, DashboardStats, DailyReport, MonthlyReport, SetupInput, CloudDriveInfo } from './api'
import type { Patient, PatientTreatment, Appointment, Invoice, Payment, Treatment, ClinicSettings, BackupInfo } from './models'

export {}

declare global {
  interface Window {
    go: {
      handler: {
        AuthHandler: {
          Login(username: string, password: string): Promise<AuthResponse>
          Logout(): Promise<void>
          GetCurrentUser(): Promise<AuthResponse>
          ChangePassword(oldPassword: string, newPassword: string): Promise<void>
        }
        PatientHandler: {
          CreatePatient(input: CreatePatientInput): Promise<Patient>
          UpdatePatient(id: string, input: CreatePatientInput): Promise<Patient>
          GetPatient(id: string): Promise<Patient>
          ListPatients(page: number, pageSize: number, search: string): Promise<PatientListResponse>
          DeletePatient(id: string): Promise<void>
          GetPatientHistory(patientID: string): Promise<PatientTreatment[]>
          CheckDuplicatePhone(phone: string): Promise<Patient | null>
          GetPatientCount(): Promise<number>
        }
        AppointmentHandler: {
          CreateAppointment(input: CreateAppointmentInput): Promise<Appointment>
          UpdateAppointment(id: string, input: CreateAppointmentInput): Promise<Appointment>
          CancelAppointment(id: string, reason: string): Promise<void>
          CompleteAppointment(id: string): Promise<void>
          GetTodayAppointments(): Promise<Appointment[]>
          GetAppointmentsByDate(date: string): Promise<Appointment[]>
          GetWeekAppointments(startDate: string, endDate: string): Promise<Appointment[]>
          GetPatientAppointments(patientID: string): Promise<Appointment[]>
          GetAppointment(id: string): Promise<Appointment>
        }
        InvoiceHandler: {
          CreateInvoice(input: CreateInvoiceInput): Promise<Invoice>
          GetInvoice(id: string): Promise<Invoice>
          ListInvoices(page: number, pageSize: number, status: string, startDate: string, endDate: string, patientID: string, search: string): Promise<InvoiceListResponse>
          RecordPayment(input: RecordPaymentInput): Promise<Payment>
          VoidInvoice(id: string, reason: string): Promise<void>
          GetPatientOutstanding(patientID: string): Promise<number>
          GetPatientInvoices(patientID: string): Promise<Invoice[]>
        }
        DashboardHandler: {
          GetDashboardStats(): Promise<DashboardStats>
          GetDailyReport(date: string): Promise<DailyReport>
          GetMonthlyReport(year: number, month: number): Promise<MonthlyReport>
        }
        SettingsHandler: {
          IsSetupComplete(): Promise<boolean>
          CompleteSetup(input: SetupInput): Promise<void>
          GetClinicSettings(): Promise<ClinicSettings>
          UpdateClinicSettings(settings: ClinicSettings): Promise<void>
          UploadLogo(base64Data: string): Promise<void>
          RemoveLogo(): Promise<void>
          ListTreatments(): Promise<Treatment[]>
          ListAllTreatments(): Promise<Treatment[]>
          CreateTreatment(name: string, code: string, category: string, description: string, defaultPrice: number): Promise<Treatment>
          UpdateTreatment(id: string, name: string, code: string, category: string, description: string, defaultPrice: number): Promise<void>
          DeleteTreatment(id: string): Promise<void>
        }
        BackupHandler: {
          CreateBackup(destinationDir: string): Promise<BackupInfo>
          CreateCloudBackup(): Promise<BackupInfo>
          DetectCloudDrives(): Promise<CloudDriveInfo[]>
          RestoreFromBackup(backupPath: string): Promise<void>
          VerifyBackup(filePath: string): Promise<boolean>
          ListBackups(): Promise<BackupInfo[]>
          GetAutoBackupPath(): Promise<string>
        }
      }
    }
  }
}
