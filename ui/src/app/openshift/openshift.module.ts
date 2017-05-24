import {NgModule} from '@angular/core';
import {CommonModule} from '@angular/common';
import {EditQuotasComponent} from './edit-quotas/edit-quotas.component';
import {OpenshiftService} from './openshift.service';
import {CoreModule} from '../core/core.module';

@NgModule({
  imports: [
    CommonModule,
    CoreModule
  ],
  providers: [OpenshiftService],
  declarations: [EditQuotasComponent]
})
export class OpenshiftModule {
}
