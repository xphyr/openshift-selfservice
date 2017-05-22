import {Injectable} from '@angular/core';
import {Http, Response, Headers, RequestOptions} from "@angular/http";
import {NotificationsService} from "angular2-notifications";
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import {Observable} from "rxjs/Observable";
import 'rxjs/add/observable/throw'
import {Router} from "@angular/router";
import {Subject} from "rxjs/Subject";

@Injectable()
export class AuthService {
  public userData;

  constructor(private http: Http, private notificationService: NotificationsService,
              private router: Router) {

    let existingUserData = localStorage.getItem("AUTH");
    if (existingUserData) {
      this.userData = JSON.parse(existingUserData);
    }
  }

  login(username: string, password: string) {
    this.http.post('/login', {
      username,
      password
    })
      .map((res: Response) => res.json())
      .catch(this.handleError.bind(this))
      .subscribe(data => {
        this.handleToken(data)
      });
  }

  isAuthenticated(): Observable<boolean> {
    let subject = new Subject<boolean>();
    if (!this.userData) {
      subject.next(false);
    } else {
      // Check token on backend
      this.http.get('/auth/check_token', this.getAuthHeaders())
        .subscribe((r: Response) => {
          if (r.status != 200) {
            console.log('Token no longer valid, new login needed');
            subject.next(false);
          } else {
            subject.next(true);
          }
        });
    }
    return subject.asObservable();
  }

  getAuthHeaders(): RequestOptions {
    if (this.userData && this.userData.token) {
      let headers = new Headers();
      headers.set('Authorization', 'Bearer ' + this.userData.token);
      return new RequestOptions({headers});
    }
  }

  private handleToken(data) {
    // Decode JWT
    let base64Url = data.token.split('.')[1];
    let base64 = base64Url.replace('-', '+').replace('_', '/');
    this.userData = {
      token: data.token,
      user: JSON.parse(window.atob(base64))
    };
    localStorage.setItem('AUTH', JSON.stringify(this.userData));
    this.notificationService.success("Login erfolgreich")
    this.router.navigate(['/home']);
  }

  private handleError(error: Response | any) {
    let msg: string;
    if (error instanceof Response) {
      msg = error.json().message;
    } else {
      msg = error.message ? error.message : error.toString();
    }
    this.notificationService.error("Fehler beim Login", msg);
    return Observable.throw(msg);
  }
}
