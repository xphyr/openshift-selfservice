import { Injectable } from '@angular/core';
import {ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot} from '@angular/router';
import { AuthService } from './auth.service';
import {Subject} from "rxjs/Subject";
import 'rxjs/add/operator/first';

@Injectable()
export class CanActivateViaAuthGuard implements CanActivate {

  constructor(private authService: AuthService, private router: Router) {}

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot) {
    var authenticated = this.authService.isAuthenticated();
    var subject = new Subject<boolean>();
    authenticated.subscribe(
      (res) => {
        if(!res && state.url !== '/login') {
          console.log("redirecting to login")
          this.router.navigate(['/login']);
        }
        subject.next(res);
      });
    return subject.asObservable().first();
  }
}
