import { OseSelfservicePage } from './app.po';

describe('ose-selfservice App', () => {
  let page: OseSelfservicePage;

  beforeEach(() => {
    page = new OseSelfservicePage();
  });

  it('should display message saying app works', () => {
    page.navigateTo();
    expect(page.getParagraphText()).toEqual('app works!');
  });
});
