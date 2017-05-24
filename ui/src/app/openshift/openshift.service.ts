import {Injectable} from '@angular/core';
import {Http} from '@angular/http';
import {AuthService} from '../core/auth/auth.service';
import {Observable} from "rxjs/Observable";

@Injectable()
export class OpenshiftService {

  constructor(private authService: AuthService, private http: Http) {
  }

  updateQuotas(projectName: string, cpu: number, memory: number): Observable<any> {
    return this.http.post('/auth/openshift/editquotas', {
      projectName,
      cpu,
      memory
    }, this.authService.getAuthHeaders());
  }
}
