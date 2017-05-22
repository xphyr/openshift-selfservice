import {Component} from '@angular/core';
import {AuthService} from "../auth/auth.service";

@Component({
    selector: 'navbar',
    templateUrl: './nav.component.html',
    styleUrls: ['./nav.component.css']
})
export class NavComponent {
    constructor(public authService: AuthService) {
    }
}
