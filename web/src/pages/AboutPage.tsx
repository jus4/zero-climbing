import { Divider } from "../components/ui/Divider";
import { Heading } from "../components/ui/Heading";
import { Paragraph } from "../components/ui/Paragraph";

export function AboutPage() {
  return (
    <div className="container mx-auto max-w-3xl px-4 py-10">
      <Heading as="h1" size="xl" align="center">
        Miksi ?
      </Heading>

      <Heading as="h2" size="lg" className="mt-6 mb-4">
        Onko järkeä käydä läpi paikkoja missä muut on jo käyneet ?
      </Heading>
      <Paragraph size="lg">
        Ei ole. Kiipeily paikkojen kartoittamisessa tulee aina mieleen kysymys onko täällä joku
        käynyt. Hyvin mahdollista että on mutta pakko käydä vielä katosomassa varman päälle.
        Palvelun tarkoitus on auttaa kiipeily yhteisöä kartoittamaan lisää mahdollisia kiipeily
        paikkoja jakamalla informaatiota missä kiipeilyä ei löytynyt. Kartalla paikat voi näyttää
        lupaavilta mutta todellisuudessa vastaan tulee vaan pieniä kiviä 1h metsässä vaeltamisen
        jälkeen.
      </Paragraph>

      <Divider />

      <Heading as="h2" size="lg" className="mb-4">
        Tarkoitus
      </Heading>
      <Paragraph size="lg">
        Palvelun tarkoitus on auttaa uusien kiipeily alueiden kartoittamisessa jakamalla tietoa
        missä kiipeilyä ole. Voit myös jakaa alueita missä on potenttiaali mutta kiinnostus tai aika
        ei vaan riitä alueen kehittämiseen. Jos haluat voit myös lisätä omat salapaikat kartalle
        ettei sinne kukaan muu satu pääsemään jos siltä tuntuu. Voit myös pitää yllä kartassa omia
        alueita niin että muut ei niitä löydä jos niin haluat.
      </Paragraph>

      <Divider />

      <Heading as="h2" size="lg" className="mb-4">
        Miten käytetään
      </Heading>
      <Paragraph size="lg">
        Palveluun voi tuoda valmista lokaatio informaatiota Maastokartta sovelluksen kautta muita
        sovelluksia tuetaan ehkä jatksossa tai mahdollista csv importtausta. Palvelu vaatii
        kirjautumisen gmail tilin kautta. Kun olet tuonut datan voit valita mitkä lokaatiot ovat
        julkisesti näkyvillä ja mitkä ei.
      </Paragraph>
    </div>
  );
}
