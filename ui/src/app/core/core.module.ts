import {CommonModule} from '@angular/common';
import {NgModule, Optional, SkipSelf} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {HttpModule} from '@angular/http';
import {RouterModule} from '@angular/router';
import {throwIfAlreadyLoaded} from './module-import-guard';

import {NavComponent} from './nav/nav.component';
import {AuthService} from "./auth/auth.service";
import {BrowserAnimationsModule} from "@angular/platform-browser/animations";
import {CanActivateViaAuthGuard} from "./auth/auth.guard";

@NgModule({
  imports: [
    CommonModule,
    FormsModule,
    HttpModule,
    RouterModule,
    BrowserAnimationsModule
  ],
  providers: [AuthService, CanActivateViaAuthGuard],
  declarations: [NavComponent],
  exports: [NavComponent, FormsModule, BrowserAnimationsModule]
})
export class CoreModule {
  constructor(@Optional() @SkipSelf() parentModule: CoreModule) {
    throwIfAlreadyLoaded(parentModule, 'core module');
  }
}
