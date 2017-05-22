import {Component} from '@angular/core';

interface NavItem {
    displayName: string,
    routerLink: string
}

@Component({
    selector: 'navbar',
    templateUrl: './nav.component.html',
    styleUrls: ['./nav.component.css']
})
export class NavComponent {

    public navItems: Array<NavItem> = [
        {displayName: 'Home', routerLink: ''},
        {displayName: 'About', routerLink: 'about'},
        {displayName: 'Theme', routerLink: 'theme'}
    ]

    constructor() {
    }
}
