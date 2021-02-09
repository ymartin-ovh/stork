import { async, ComponentFixture, TestBed } from '@angular/core/testing'
import { SimpleChange } from '@angular/core'
import { FormsModule } from '@angular/forms'
import { By } from '@angular/platform-browser'

import { RenameAppDialogComponent } from './rename-app-dialog.component'

describe('RenameAppDialogComponent', () => {
    let component: RenameAppDialogComponent
    let fixture: ComponentFixture<RenameAppDialogComponent>

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [FormsModule],
            declarations: [RenameAppDialogComponent],
        }).compileComponents()
    }))

    beforeEach(() => {
        fixture = TestBed.createComponent(RenameAppDialogComponent)
        component = fixture.componentInstance
        fixture.detectChanges()
    })

    it('should create', () => {
        expect(component).toBeTruthy()
        expect(component.visible).toBeFalse()
    })

    it('should submit valid app name', () => {
        component.appId = 2
        component.visible = true
        fixture.detectChanges()

        const appNameInput = fixture.debugElement.query(By.css('#app-name-input'))
        const appNameInputElement = appNameInput.nativeElement

        // Set valid app name and emit the events required to propagate this
        // app name to the component level.
        appNameInputElement.value = 'dhcp-server-floor1'
        appNameInputElement.dispatchEvent(new Event('input'))
        appNameInputElement.dispatchEvent(new KeyboardEvent('keyup'))
        fixture.detectChanges()

        // Make sure the new name was propagated and that the new value was
        // properly validated.
        expect(component.appName).toBe('dhcp-server-floor1')
        expect(appNameInput.classes.hasOwnProperty('ng-valid')).toBeTrue()
        expect(appNameInput.classes['ng-valid']).toBeTrue()

        spyOn(component.submitted, 'emit')
        spyOn(component.cancelled, 'emit')

        // Submit the new name.
        component.save()

        // The emitter indicating that the name was submitted should have been
        // triggered. The one that indicates cancelling the operation should
        // not.
        expect(component.submitted.emit).toHaveBeenCalled()
        expect(component.submitted.emit).toHaveBeenCalledWith('dhcp-server-floor1')
        expect(component.cancelled.emit).not.toHaveBeenCalled()
    })

    it('should validate app name with double at character', () => {
        component.appId = 2
        component.visible = true
        // Simulate the case that there are two machines in the system.
        component.existingMachines = new Set(['machine1', 'machine2'])
        fixture.detectChanges()

        const appNameInput = fixture.debugElement.query(By.css('#app-name-input'))
        const appNameInputElement = appNameInput.nativeElement

        // Set app name with two "at" characters. The check for referenced machine
        // name should be skipped.
        appNameInputElement.value = 'fix@@machine3'
        appNameInputElement.dispatchEvent(new Event('input'))
        appNameInputElement.dispatchEvent(new KeyboardEvent('keyup'))
        fixture.detectChanges()

        // Make sure the new name was propagated and that the new value was
        // properly validated.
        expect(component.appName).toBe('fix@@machine3')
        expect(appNameInput.classes.hasOwnProperty('ng-valid')).toBeTrue()
        expect(appNameInput.classes['ng-valid']).toBeTrue()
    })

    it('should cancel rename', () => {
        component.appId = 2
        component.visible = true
        component.appName = 'first'
        // Need to call this directly to ensure that the original app name
        // was saved.
        component.ngOnChanges({
            appId: new SimpleChange(null, component.appId, false),
        })
        fixture.detectChanges()

        const appNameInput = fixture.debugElement.query(By.css('#app-name-input'))
        const appNameInputElement = appNameInput.nativeElement

        // Set the new value to an empty string.
        appNameInputElement.value = ''
        appNameInputElement.dispatchEvent(new Event('input'))
        appNameInputElement.dispatchEvent(new KeyboardEvent('keyup'))
        fixture.detectChanges()

        // Make sure it is set to an empty string and that it is treated
        // as an invalid name.
        expect(component.appName.length).toBe(0)
        expect(appNameInput.classes.hasOwnProperty('ng-invalid')).toBeTrue()
        expect(appNameInput.classes['ng-invalid']).toBeTrue()

        spyOn(component.submitted, 'emit')
        spyOn(component.cancelled, 'emit')

        // Cancel the rename. Appropriate emitter should be triggered.
        component.cancel()
        expect(component.submitted.emit).not.toHaveBeenCalled()
        expect(component.cancelled.emit).toHaveBeenCalled()

        // The original name should be restored and the error messages should
        // be cleared.
        expect(component.appName).toBe('first')
        expect(component.invalid).toBeFalse()
        expect(component.errorText.length).toBe(0)
    })

    it('should reject a name belonging to another app', () => {
        component.appId = 2
        component.visible = true
        component.appName = 'dhcp-server-floor2'
        // Simulate the case that there are two apps defined in the system. One
        // of them is our app.
        component.existingApps = new Map([
            ['dhcp-server-floor1', 1],
            ['dhcp-server-floor2', 2],
        ])
        fixture.detectChanges()

        const appNameInput = fixture.debugElement.query(By.css('#app-name-input'))
        const appNameInputElement = appNameInput.nativeElement

        // Rename our app to the name of the other app.
        appNameInputElement.value = 'dhcp-server-floor1'
        appNameInputElement.dispatchEvent(new Event('input'))
        appNameInputElement.dispatchEvent(new KeyboardEvent('keyup'))
        fixture.detectChanges()

        // Make sure the change was propagated to the component level and that
        // this name is treated as invalid one.
        expect(component.appName).toBe('dhcp-server-floor1')
        expect(appNameInput.classes.hasOwnProperty('ng-invalid')).toBeTrue()
        expect(appNameInput.classes['ng-invalid']).toBeTrue()

        // Ensure that the appropriate error message is displayed.
        const appNameInputHelp = fixture.debugElement.query(By.css('#app-name-input-help'))
        expect(appNameInputHelp.nativeElement.innerText).toBe('App with this name already exists.')

        // Ensure that the submit button is disabled.
        const renameButton = fixture.debugElement.query(By.css('#rename-button'))
        expect(renameButton.properties.hasOwnProperty('disabled')).toBeTrue()
        expect(renameButton.properties.disabled).toBeTrue()
    })

    it('should reject a name referencing non-existing machine', () => {
        component.appId = 2
        component.visible = true
        // Simulate the case that there are two machines in the system.
        component.existingMachines = new Set(['machine1', 'machine2'])
        fixture.detectChanges()

        const appNameInput = fixture.debugElement.query(By.css('#app-name-input'))
        const appNameInputElement = appNameInput.nativeElement

        // The new app name references machine3 which doesn't exist.
        appNameInputElement.value = 'lion@machine3'
        appNameInputElement.dispatchEvent(new Event('input'))
        appNameInputElement.dispatchEvent(new KeyboardEvent('keyup', { key: 'n' }))
        fixture.detectChanges()

        // Make sure the name has been propagated to the component level.
        expect(component.appName).toBe('lion@machine3')
        // Make sure this name doesn't validate because the given machine
        // does not exist.
        expect(appNameInput.classes.hasOwnProperty('ng-invalid')).toBeTrue()
        expect(appNameInput.classes['ng-invalid']).toBeTrue()

        // Make sure that the error message is displayed.
        const appNameInputHelp = fixture.debugElement.query(By.css('#app-name-input-help'))
        expect(appNameInputHelp.nativeElement.innerText).toBe('Machine machine3 does not exist.')
    })

    it('should reject an empty name', () => {
        component.appId = 2
        component.visible = true
        fixture.detectChanges()

        const appNameInput = fixture.debugElement.query(By.css('#app-name-input'))
        const appNameInputElement = appNameInput.nativeElement

        // App name consisting of a whitespace is treated as an empty name.
        appNameInputElement.value = '  '
        appNameInputElement.dispatchEvent(new Event('input'))
        appNameInputElement.dispatchEvent(new KeyboardEvent('keyup', { key: 'n' }))
        fixture.detectChanges()

        expect(appNameInput.classes.hasOwnProperty('ng-invalid')).toBeTrue()
        expect(appNameInput.classes['ng-invalid']).toBeTrue()

        const appNameInputHelp = fixture.debugElement.query(By.css('#app-name-input-help'))
        expect(appNameInputHelp.nativeElement.innerText).toBe('App name must not be empty.')
    })
})